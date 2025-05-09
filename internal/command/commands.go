package command

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bbarrington0099/Gator/internal/database"
	"github.com/bbarrington0099/Gator/internal/state"
	"github.com/google/uuid"
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