package userconfig

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/utils"
)

type Service struct {
	commandsRepository CommandsRepository
	filesRepository    FilesRepository
	filesystem         Filesystem
	usingConsole       string
}

func NewService(commandsRepository CommandsRepository, filesRepository FilesRepository, filesystem Filesystem, usingConsole string) *Service {
	return &Service{
		commandsRepository: commandsRepository,
		filesRepository:    filesRepository,
		filesystem:         filesystem,
		usingConsole:       usingConsole,
	}
}

func (s Service) CreateDefaultUserConfig() *entities.UserConfig {
	return &entities.UserConfig{
		UsingConsole: s.usingConsole,
	}
}

func (s Service) GetUserConfig() (*entities.UserConfig, error) {
	result, err := s.commandsRepository.GetCommands()
	if err != nil {
		return nil, err
	}
	return &entities.UserConfig{
		UsingConsole: s.usingConsole,
		Commands:     result,
	}, nil
}

func (s Service) clearFiles() error {
	err := s.filesRepository.DeleteAllFiles()
	if err != nil {
		return err
	}
	return s.filesystem.ClearFiles()
}

func (s Service) SetUserConfig(newConfig *entities.UserConfig) error {
	utils.SetDefaultCommandsNames(newConfig.Commands)
	err := s.commandsRepository.SetCommands(newConfig.Commands)
	if err != nil {
		return err
	}
	err = s.clearFiles()
	if err != nil {
		return err
	}
	return nil
}
