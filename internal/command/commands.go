package command

import (
	"context"
	"fmt"
	"os"
	"time"
	
	"github.com/google/uuid"

	"github.com/bbarrington0099/Gator/internal/database"
	"github.com/bbarrington0099/Gator/internal/state"
	"github.com/bbarrington0099/Gator/internal/RSS"
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

func HandlerLogin(s *state.State, cmd Command) (err error) {
	if len(cmd.Args) == 0 {
		err = fmt.Errorf("missing username")
		return
	}

	username := cmd.Args[0]

	_, err = s.DB.GetUser(context.Background(), username)
	if err != nil {
		fmt.Printf("User %s not found\n", username)
		os.Exit(1)
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
		fmt.Printf("Error deleting all users: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Deleted all users\n")
	return
}

func HandlerUsers(s *state.State, cmd Command) (err error) {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Error getting users: %v\n", err)
		os.Exit(1)
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

func HandlerAgg(s *state.State, cmd Command) error {
	feedURL := "https://www.wagslane.dev/index.xml"
	
	ctx := context.Background()
	feed, err := rss.FetchFeed(ctx, feedURL)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	// Simply print the struct using the default fmt formatting
	fmt.Printf("%+v\n", *feed)

	return nil
}

func HandlerAddFeed(s *state.State, cmd Command) (err error) {
	if len(cmd.Args) < 2 {
		err = fmt.Errorf("missing feed name or URL")
		return
	}

	feedName := cmd.Args[0]
	feedURL := cmd.Args[1]

	currentUser := s.Config.Current_user_name
	userInfo, err := s.DB.GetUser(context.Background(), currentUser)
	if err != nil {
		fmt.Printf("Error getting user ID: %v\n", err)
		os.Exit(1)
	}
	currentUserID := userInfo.ID

	newFeedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedURL,
		UserID:    currentUserID,
	}

	context := context.Background()

	newFeed, err := s.DB.CreateFeed(context, newFeedParams)
	if err != nil {
		return
	}

	fmt.Printf("%+v\n", newFeed)
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