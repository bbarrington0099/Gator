package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/bbarrington0099/Gator/internal/command"
	"github.com/bbarrington0099/Gator/internal/config"
	"github.com/bbarrington0099/Gator/internal/database"
	"github.com/bbarrington0099/Gator/internal/state"
	_ "github.com/lib/pq"
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

	db, err := sql.Open("postgres", current_config.Db_url)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bdQueries := database.New(db)

	state.DB = bdQueries

	commands := command.Commands{}
	commands.Register("login", command.HandlerLogin)
	commands.Register("register", command.HandlerRegister)

	if _, ok := commands.ExecutableCommand[os.Args[1]]; !ok {
		log.Fatalf("unknown command: %s", os.Args[1])
	}

	cmd := command.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	err = commands.Run(&state, cmd)
	if err != nil {
		log.Fatal(err)
	}
}
