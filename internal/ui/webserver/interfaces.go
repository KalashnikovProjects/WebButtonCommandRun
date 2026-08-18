package webserver

import (
	"context"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
)

type Runner interface {
	RunCommand(ctx context.Context, commandId uint, options entities.TerminalOptions) (*entities.CommandInputOutput, error)
}

type Commands interface {
	DefaultCommand() *entities.Command
	AppendCommand(command *entities.Command) error
	DeleteCommand(commandId uint) error
	PatchCommand(commandId uint, newCommand *entities.Command) error
	PutCommand(commandId uint, newCommand *entities.Command) error
	GetCommandsList() ([]entities.Command, error)
	GetCommand(commandId uint) (*entities.Command, error)
	CommandExists(commandId uint) (bool, error)
}

type Files interface {
	AppendFile(commandID uint, fileBytes []byte, data *entities.FileParams) error
	DeleteFile(commandId, fileId uint) error
	PatchFile(commandId, fileId uint, newFile *entities.EmbeddedFile) error
	PutFile(commandId, fileId uint, newFile *entities.EmbeddedFile) error
	GetFile(commandId, fileId uint) (*entities.EmbeddedFile, error)
	GetCommandFiles(commandId uint) ([]entities.EmbeddedFile, error)
	GetAllFilesList() ([]entities.EmbeddedFile, error)
	DownloadFile(commandId, fileId uint) (*entities.EmbeddedFile, []byte, error)
	DownloadCommandFilesInArchive(commandId uint) ([]byte, error)
	DownloadAllFilesInArchive() ([]byte, error)
	ImportAllFilesFromZipArchive(data []byte) error
}

type UserConfig interface {
	CreateDefaultUserConfig() *entities.UserConfig
	GetUserConfig() (*entities.UserConfig, error)
	SetUserConfig(newConfig *entities.UserConfig) error
}
