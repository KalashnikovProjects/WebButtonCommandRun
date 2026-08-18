package commands

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/utils"
)

type Service struct {
	commandsRepository   CommandsRepository
	defaultCommandRunDir string
}

func NewService(commandsRepository CommandsRepository, defaultCommandRunDir string) *Service {
	return &Service{
		commandsRepository:   commandsRepository,
		defaultCommandRunDir: defaultCommandRunDir,
	}
}

func (s Service) DefaultCommand() *entities.Command {
	return &entities.Command{
		Dir: s.defaultCommandRunDir,
	}
}

func (s Service) AppendCommand(command *entities.Command) error {
	utils.SetDefaultCommandsName(command)
	if err := utils.CheckName(command.Name); err != nil {
		return err
	}
	return s.commandsRepository.AppendCommand(command)
}

func (s Service) DeleteCommand(commandId uint) error {
	return s.commandsRepository.DeleteCommand(commandId)
}

func (s Service) PatchCommand(commandId uint, newCommand *entities.Command) error {
	if newCommand.Name != "" {
		if err := utils.CheckName(newCommand.Name); err != nil {
			return err
		}
	}
	return s.commandsRepository.PatchCommand(commandId, newCommand)
}

func (s Service) PutCommand(commandId uint, newCommand *entities.Command) error {
	utils.SetDefaultCommandsName(newCommand)
	if err := utils.CheckName(newCommand.Name); err != nil {
		return err
	}
	return s.commandsRepository.PutCommand(commandId, newCommand)
}

func (s Service) GetCommandsList() ([]entities.Command, error) {
	return s.commandsRepository.GetCommands()
}

func (s Service) GetCommand(commandId uint) (*entities.Command, error) {
	return s.commandsRepository.GetCommand(commandId)
}

func (s Service) CommandExists(commandId uint) (bool, error) {
	return s.commandsRepository.CommandExists(commandId)
}
