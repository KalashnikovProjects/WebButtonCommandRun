package runner

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
)

type Runner interface {
	RunCommand(command string, options entities.TerminalOptions) (entities.RunningCommand, error)
}

type CommandsRepository interface {
	GetCommand(id uint) (*entities.Command, error)
}

type FilesRepository interface {
	GetCommandFiles(commandId uint) ([]entities.EmbeddedFile, error)
}
