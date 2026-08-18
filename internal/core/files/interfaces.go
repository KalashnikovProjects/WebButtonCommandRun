package files

import "github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"

type CommandsRepository interface {
	CommandExists(id uint) (bool, error)
}

type FilesRepository interface {
	AppendFile(file *entities.EmbeddedFile) error
	UpdateFile(commandId, id uint, new *entities.EmbeddedFile) error
	PatchFile(commandId, id uint, new *entities.EmbeddedFile) error
	DeleteFile(commandId, id uint) error
	GetFile(commandId, id uint) (*entities.EmbeddedFile, error)
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
