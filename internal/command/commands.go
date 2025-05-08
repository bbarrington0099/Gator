package command

import (
	"fmt"

	"github.com/bbarrington0099/Gator/internal/state"
)

type Commands struct {
	executableCommand map[string]func(*state.State, Command) error
}

func (c *Commands) Run(s *state.State, cmd Command) (err error) {
	if commandFunc, ok := c.executableCommand[cmd.Name]; ok {
		err = commandFunc(s, cmd)
	} else {
		err = fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return
}

func (c *Commands) Register(name string, f func(*state.State, Command) error) {
	if c.executableCommand == nil {
		c.executableCommand = make(map[string]func(*state.State, Command) error)
	}
	c.executableCommand[name] = f
}

func HandlerLogin(s *state.State, cmd Command) (err error) {
	if len(cmd.Args) == 0 {
		err = fmt.Errorf("missing username")
		return
	}

	err = s.Config.SetUser(cmd.Args[0])
	if err != nil {
		return
	}

	fmt.Printf("Logged in as %s\n", cmd.Args[0])
	return
}