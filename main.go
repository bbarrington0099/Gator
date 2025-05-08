package main

import (
	"log"
	"os"

	"github.com/bbarrington0099/Gator/internal/command"
	"github.com/bbarrington0099/Gator/internal/config"
	"github.com/bbarrington0099/Gator/internal/state"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("missing command")
	}

	current_config, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	state := state.State{
		Config: &current_config,
	}

	commands := command.Commands{}
	commands.Register("login", command.HandlerLogin)

	cmd := command.Command{
		Name: "login",
		Args: os.Args[2:],
	}

	err = commands.Run(&state, cmd)
	if err != nil {
		log.Fatal(err)
	}
}