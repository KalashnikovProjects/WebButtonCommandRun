package commands

import "github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"

type CommandsRepository interface {
	AppendCommand(command *entities.Command) error
	DeleteCommand(id uint) error
	GetCommands() ([]entities.Command, error)
	SetCommands(commands []entities.Command) error
	GetCommand(id uint) (*entities.Command, error)
	PutCommand(id uint, new *entities.Command) error
	PatchCommand(id uint, new *entities.Command) error
	CommandExists(id uint) (bool, error)
}
