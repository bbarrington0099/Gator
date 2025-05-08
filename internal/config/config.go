package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	Db_url string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func Read() (gator_config Config, err error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return
	}

	config_data, err := os.ReadFile(configFilePath)
	if err != nil {
		return
	}

	err = json.Unmarshal(config_data, &gator_config)
	
	return
}

func (config Config) SetUser(user_name string) (err error) {
	config.Current_user_name = user_name
	return write(config)
}

func getConfigFilePath() (path string, err error) {
	home_dir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	path = filepath.Join(home_dir, configFileName)
	_, err = os.Stat(path);

	return
}

func write(updated_config Config) (err error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return
	}

	config_data, err := json.Marshal(updated_config)
	if err != nil {
		return
	}

	err = os.WriteFile(configFilePath, config_data, 0644)
	if err != nil {
		return
	}

	return
}