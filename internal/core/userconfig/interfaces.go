package userconfig

import "github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"

type CommandsRepository interface {
	GetCommands() ([]entities.Command, error)
	SetCommands(commands []entities.Command) error
}

type FilesRepository interface {
	DeleteAllFiles() error
}

type Filesystem interface {
	ClearFiles() error
}
