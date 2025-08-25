package data

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
)

func (s service) GetUserConfig() (entities.UserConfig, error) {
	commands, err := s.commandsRepo.GetCommands()
	if err != nil {
		return entities.UserConfig{}, err
	}
	return entities.UserConfig{
		UsingConsole: config.Config.Console,
		Commands:     commands,
	}, nil
}

func (s service) SetUserConfig(newConfig entities.UserConfig) error {
	SetDefaultCommandsNames(newConfig.Commands)
	err := s.commandsRepo.SetCommands(newConfig.Commands)
	if err != nil {
		return err
	}
	err = s.clearFiles()
	if err != nil {
		return err
	}
	return nil
}
