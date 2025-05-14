package command

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	"strconv"

	"github.com/google/uuid"

	"github.com/bbarrington0099/Gator/internal/RSS"
	"github.com/bbarrington0099/Gator/internal/database"
	"github.com/bbarrington0099/Gator/internal/state"
)

type Commands struct {
	ExecutableCommand map[string]func(*state.State, Command) error
}

func (c *Commands) Run(s *state.State, cmd Command) (err error) {
	if commandFunc, ok := c.ExecutableCommand[cmd.Name]; ok {
		err = commandFunc(s, cmd)
	} else {
		err = fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return
}

func (c *Commands) Register(name string, f func(*state.State, Command) error) {
	if c.ExecutableCommand == nil {
		c.ExecutableCommand = make(map[string]func(*state.State, Command) error)
	}
	c.ExecutableCommand[name] = f
}

// Command Handlers

func HandlerLogin(s *state.State, cmd Command) (err error) {
	if len(cmd.Args) == 0 {
		err = fmt.Errorf("missing username")
		return
	}

	username := cmd.Args[0]

	_, err = s.DB.GetUser(context.Background(), username)
	if err != nil {
		return
	}

	err = s.Config.SetUser(username)
	if err != nil {
		return
	}

	fmt.Printf("Logged in as %s\n", username)
	return
}

func HandlerRegister(s *state.State, cmd Command) (err error) {
	if len(cmd.Args) == 0 {
		err = fmt.Errorf("missing username")
		return
	}

	newUserParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	}

	context := context.Background()

	newUser, err := s.DB.CreateUser(context, newUserParams)
	if err != nil {
		return
	}

	s.Config.SetUser(newUser.Name)

	fmt.Printf("Registered new user %s\n\tID: %v\n\tCreatedAt: %v\n\tUpdatedAt: %v\n", newUser.Name, newUser.ID, newUser.CreatedAt, newUser.UpdatedAt)
	return
}

func HandlerReset(s *state.State, cmd Command) (err error) {
	err = s.DB.DeleteAllUsers(context.Background())
	if err != nil {
		return
	}
	fmt.Printf("Deleted all users\n")
	return
}

func HandlerUsers(s *state.State, cmd Command) (err error) {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return
	}

	for _, user := range users {
		currentUser := s.Config.Current_user_name
		if user.Name == currentUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return
}

func HandlerAgg(s *state.State, cmd Command) (err error) {
	if len(cmd.Args) < 1 {
		err = fmt.Errorf("missing time between requests")
		return
	}

	time_between_reqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return
	}

	fmt.Printf("Collecting feeds every %s\n", time_between_reqs)

	ticker := time.NewTicker(time_between_reqs)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			return
		}
	}
}

func HandlerAddFeed(s *state.State, cmd Command, user database.User) (err error) {
	if len(cmd.Args) < 2 {
		err = fmt.Errorf("missing feed name or URL")
		return
	}

	feedName := cmd.Args[0]
	feedURL := cmd.Args[1]

	newFeedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedURL,
		UserID:    user.ID,
	}

	context := context.Background()

	newFeed, err := s.DB.CreateFeed(context, newFeedParams)
	if err != nil {
		return
	}

	fmt.Printf("%+v\n", newFeed)

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    newFeed.ID,
	}

	feedFollow, err := s.DB.CreateFeedFollow(context, feedFollowParams)
	if err != nil {
		fmt.Printf("Error creating feed follow: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s now follows %s\n", feedFollow.UserName, feedFollow.FeedName)

	return
}

func HandlerFeeds(s *state.State, cmd Command) (err error) {
	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		fmt.Printf("Error getting feeds: %v\n", err)
		os.Exit(1)
	}

	for _, feed := range feeds {
		fmt.Printf("%s created * %s (%s)\n", feed.UserName, feed.Name, feed.Url)
	}
	return
}

func HandlerFollow(s *state.State, cmd Command, user database.User) (err error) {
	if len(cmd.Args) < 1 {
		err = fmt.Errorf(("missing url"))
		return
	}

	feedURL := cmd.Args[0]

	feed, err := s.DB.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		fmt.Printf("Error getting feed: %v\n", err)
		os.Exit(1)
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	feedFollow, err := s.DB.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		fmt.Printf("Error creating feed follow: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s %s\n", feedFollow.FeedName, feedFollow.UserName)

	return
}

func HandlerFollowing(s *state.State, cmd Command, user database.User) (err error) {
	if len(cmd.Args) > 0 {
		err = fmt.Errorf("too many arguments")
		return
	}

	follows, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		fmt.Printf("Error getting feed follows: %v\n", err)
		os.Exit(1)
	}

	for _, follow := range follows {
		fmt.Printf("%s %s\n", follow.FeedName, follow.UserName)
	}
	return
}

func HandlerUnfollow(s *state.State, cmd Command, user database.User) (err error) {
	if len(cmd.Args) < 1 {
		err = fmt.Errorf("missing url")
		return
	}

	feedURL := cmd.Args[0]

	feed, err := s.DB.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		fmt.Printf("Error getting feed: %v\n", err)
		os.Exit(1)
	}

	err = s.DB.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		fmt.Printf("Error deleting feed follow: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s unfollowed %s\n", user.Name, feed.Name)

	return
}

func HandlerBrowse(s *state.State, cmd Command, user database.User) (err error) {
	var limit int
	
	if len(cmd.Args) < 1 {
		limit = 2
	} else {
		parsedLimit, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %v", err)
		}
		limit = parsedLimit
	}
	
	userPostsParams := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	posts, err := s.DB.GetPostsForUser(context.Background(), userPostsParams)
	if err != nil {
		return
	} 
	
	for _, post := range posts {
		fmt.Printf("%s %s\n %s\n", post.FeedName, post.Title, post.Description)
	}

	return
}

// Helper Functions

func scrapeFeeds(s *state.State) (err error) {
	feed, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("Error getting next feed to fetch: %v\n", err)
		os.Exit(1)
	}

	err = s.DB.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Printf("Error marking feed as fetched: %v\n", err)
		os.Exit(1)
	}

	feedData, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("Error fetching feed: %v\n", err)
		os.Exit(1)
	}

	for _, item := range feedData.Channel.Item {
		pubTime, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			pubTime, err = time.Parse(time.RFC1123, item.PubDate)
			if err != nil {
				fmt.Printf("Error parsing publication date '%s': %v\n", item.PubDate, err)
				continue
			}
		}

		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			FeedID:      feed.ID,
			Description: item.Description,
			PublishedAt: pubTime,
		}

		_, err = s.DB.CreatePost(context.Background(), postParams)
		if err != nil {
			if strings.Contains(err.Error(), "posts_url_key") {
				fmt.Printf("Post %s already exists\n", item.Link)
			} else {
				fmt.Printf("Error creating post: %v\n", err)
			}
			continue
		}
	}

	return
}

// Middleware

func MiddlewareCurrentUser(handler func(s *state.State, cmd Command, user database.User) (err error)) func (s *state.State, cmd Command) (err error) {
	return func(s *state.State, cmd Command) (err error) {
		currentUser := s.Config.Current_user_name
		userInfo, err := s.DB.GetUser(context.Background(), currentUser)
		if err != nil {
			fmt.Printf("Error getting user ID: %v\n", err)
			os.Exit(1)
		}

		err = handler(s, cmd, userInfo)
		return
	}
}