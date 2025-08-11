package usecases

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/storage"
)

type DB struct {
	DB *storage.DB
}

var ErrCommandNotFound = storage.ErrorNotFound
var ErrCantPatchStruct = errors.New("error cant patch command structure")

func CreateDB() (DB, error) {
	db, err := storage.Connect()
	if err != nil {
		return DB{}, err
	}
	return DB{&db}, nil
}

func SetDefaultCommandsNames(commands []entities.Command) {
	for i := 0; i < len(commands); i++ {
		SetDefaultCommandsName(&commands[i])
	}
}

func SetDefaultCommandsName(command *entities.Command) {
	if command.Name == "" {
		command.Name = RandomCommandName()
	}
}

func RandomCommandName() string {
	return fmt.Sprintf("Command %d", rand.Intn(100))
}

func SetDefaultFilesNames(files []entities.EmbeddedFile) {
	for i := 0; i < len(files); i++ {
		SetDefaultFileName(&files[i])
	}
}

func SetDefaultFileName(file *entities.EmbeddedFile) {
	if file.Name == "" {
		file.Name = fmt.Sprintf("File %d", rand.Intn(100))
	}
}

func RandomFileName() string {
	return fmt.Sprintf("File %d", rand.Intn(100))
}

func (db DB) GetUserConfig() (entities.UserConfig, error) {
	commands, err := db.DB.GetCommands()
	if err != nil {
		return entities.UserConfig{}, err
	}
	files, err := db.DB.GetAllFiles()
	if err != nil {
		return entities.UserConfig{}, err
	}
	return entities.UserConfig{
		UsingConsole: config.Config.Console,
		Commands:     commands,
		Files:        files,
	}, nil
}

func (db DB) SetUserConfig(newConfig entities.UserConfig) error {
	SetDefaultCommandsNames(newConfig.Commands)
	err := db.DB.SetCommands(newConfig.Commands)
	if err != nil {
		return err
	}
	err = db.DB.SetAllFiles(newConfig.Files)
	if err != nil {
		return err
	}
	return nil
}

func (db DB) AppendCommand(command entities.Command) error {
	SetDefaultCommandsName(&command)
	return db.DB.AppendCommand(command)
}

func (db DB) DeleteCommand(commandId uint) error {
	return db.DB.DeleteCommand(commandId)
}

func (db DB) PatchCommand(commandId uint, newCommand entities.Command) error {
	return db.DB.PatchCommand(commandId, newCommand)
}

func (db DB) PutCommand(commandId uint, newCommand entities.Command) error {
	SetDefaultCommandsName(&newCommand)
	return db.DB.PutCommand(commandId, newCommand)
}

func (db DB) GetCommandsList() ([]entities.Command, error) {
	return db.DB.GetCommands()
}

func (db DB) GetCommand(commandId uint) (entities.Command, error) {
	return db.DB.GetCommand(commandId)
}

// AppendFile TODO: сохранение самого файла в файловой системе
func (db DB) AppendFile(file entities.EmbeddedFile) error {
	SetDefaultFileName(&file)
	return db.DB.AppendFile(file)
}

func (db DB) DeleteFile(fileId uint) error {
	return db.DB.DeleteFile(fileId)
}

func (db DB) PatchFile(fileId uint, newFile entities.EmbeddedFile) error {
	return db.DB.PatchFile(fileId, newFile)
}

func (db DB) PutFile(fileId uint, newFile entities.EmbeddedFile) error {
	SetDefaultFileName(&newFile)
	return db.DB.UpdateFile(fileId, newFile)
}

func (db DB) GetFile(fileId uint) (entities.EmbeddedFile, error) {
	return db.DB.GetFile(fileId)
}

func (db DB) GetCommandFilesList(commandId uint) ([]entities.EmbeddedFile, error) {
	return db.DB.GetCommandFiles(commandId)
}

func (db DB) GetAllFilesList() ([]entities.EmbeddedFile, error) {
	return db.DB.GetAllFiles()
}
