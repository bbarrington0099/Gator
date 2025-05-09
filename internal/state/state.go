package state

import (
	"github.com/bbarrington0099/Gator/internal/config"
	"github.com/bbarrington0099/Gator/internal/database"
)

type State struct {
	DB *database.Queries
	Config *config.Config
}