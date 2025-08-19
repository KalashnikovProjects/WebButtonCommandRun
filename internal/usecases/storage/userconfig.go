package storage

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
)

func (db DB) GetUserConfig() (entities.UserConfig, error) {
	commands, err := db.DB.GetCommands()
	if err != nil {
		return entities.UserConfig{}, err
	}
	return entities.UserConfig{
		UsingConsole: config.Config.Console,
		Commands:     commands,
	}, nil
}

func (db DB) SetUserConfig(newConfig entities.UserConfig) error {
	SetDefaultCommandsNames(newConfig.Commands)
	err := db.DB.SetCommands(newConfig.Commands)
	if err != nil {
		return err
	}
	err = db.clearFiles()
	if err != nil {
		return err
	}
	return nil
}
