package data

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"strings"
)

type CommandsRepo interface {
	AppendCommand(command *entities.Command) error
	DeleteCommand(id uint) error
	GetCommands() ([]entities.Command, error)
	SetCommands(commands []entities.Command) error
	GetCommand(id uint) (entities.Command, error)
	PutCommand(id uint, new *entities.Command) error
	PatchCommand(id uint, new *entities.Command) error
	CommandExists(id uint) (bool, error)
}

type FilesRepo interface {
	AppendFile(file *entities.EmbeddedFile) error
	UpdateFile(commandId, id uint, new *entities.EmbeddedFile) error
	PatchFile(commandId, id uint, new *entities.EmbeddedFile) error
	DeleteFile(commandId, id uint) error
	GetFile(commandId, id uint) (entities.EmbeddedFile, error)
	GetCommandFiles(commandId uint) ([]entities.EmbeddedFile, error)
	GetCommandFilesWithCommandInfo(commandId uint) ([]entities.EmbeddedFileWithCommandInfo, error)
	GetAllFiles() ([]entities.EmbeddedFile, error)
	GetAllFilesWithCommandInfo() ([]entities.EmbeddedFileWithCommandInfo, error)
	SetAllFiles(files []entities.EmbeddedFile) error
	DeleteAllFiles() error
	SetCommandFiles(commandId uint, files []entities.EmbeddedFile) error
}

type Filesystem interface {
	SaveFile(fileId uint, bytes []byte) error
	GetFileData(fileId uint) ([]byte, error)
	DeleteFile(fileId uint) error
	ClearFiles() error
	ImportFilesFromZipArchive(data []byte) ([]entities.FileData, error)
}

type service struct {
	commandsRepo CommandsRepo
	filesRepo    FilesRepo
	filesystem   Filesystem
}

type Service interface {
	AppendCommand(command entities.Command) error
	DeleteCommand(commandId uint) error
	PatchCommand(commandId uint, newCommand entities.Command) error
	PutCommand(commandId uint, newCommand entities.Command) error
	GetCommandsList() ([]entities.Command, error)
	GetCommand(commandId uint) (entities.Command, error)
	CommandExists(commandId uint) (bool, error)

	AppendFile(commandID uint, fileBytes []byte, data entities.FileParams) error
	DeleteFile(commandId, fileId uint) error
	PatchFile(commandId, fileId uint, newFile entities.EmbeddedFile) error
	PutFile(commandId, fileId uint, newFile entities.EmbeddedFile) error
	GetFile(commandId, fileId uint) (entities.EmbeddedFile, error)
	GetCommandFilesList(commandId uint) ([]entities.EmbeddedFile, error)
	GetAllFilesList() ([]entities.EmbeddedFile, error)
	DownloadFile(commandId, fileId uint) (entities.EmbeddedFile, []byte, error)
	DownloadCommandFilesInArchive(commandId uint) ([]byte, error)
	DownloadAllFilesInArchive() ([]byte, error)
	ImportAllFilesFromZipArchive(data []byte) error
	GetUserConfig() (entities.UserConfig, error)
	SetUserConfig(newConfig entities.UserConfig) error
}

func NewService(commandsRepo CommandsRepo, filesRepo FilesRepo, filesystem Filesystem) Service {
	return &service{
		commandsRepo: commandsRepo,
		filesRepo:    filesRepo,
		filesystem:   filesystem,
	}
}

func checkName(name string) error {
	if name == "" {
		return projectErrors.ErrBadName
	}

	invalidChars := []string{
		"<", ">", ":", "\"", "|", "?", "*", "/", "\\",
		"\x00", "\x01", "\x02", "\x03", "\x04", "\x05", "\x06", "\x07",
		"\x08", "\x09", "\x0A", "\x0B", "\x0C", "\x0D", "\x0E", "\x0F",
		"\x10", "\x11", "\x12", "\x13", "\x14", "\x15", "\x16", "\x17",
		"\x18", "\x19", "\x1A", "\x1B", "\x1C", "\x1D", "\x1E", "\x1F",
	}

	result := name
	for _, char := range invalidChars {
		if strings.Contains(result, char) {
			return projectErrors.ErrBadName
		}
	}

	result = strings.TrimSpace(result)
	result = strings.Trim(result, ".")

	if result == "" {
		return projectErrors.ErrBadName
	}

	if len(result) > 255 {
		return projectErrors.ErrBadName
	}

	return nil
}
