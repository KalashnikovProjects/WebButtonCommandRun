package entities

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"io"
)

type TerminalOptions struct {
	Cols uint16   `json:"cols"`
	Rows uint16   `json:"rows"`
	Env  []string `json:"-"`
	Dir  string   `json:"-"`
}

type EmbeddedFile struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	CommandID uint   `json:"command-id" gorm:"index;->;<-:create"`
	Name      string `json:"name"`
}

type UserConfig struct {
	UsingConsole string    `json:"usingConsole"`
	Commands     []Command `json:"commands"`
}

type Command struct {
	ID      uint   `json:"id" gorm:"->;<-:create;primaryKey"`
	Name    string `json:"name"`
	Command string `json:"command"`
	Dir     string `json:"executionDir"`
}

type EmbeddedFileWithCommandInfo struct {
	EmbeddedFile
	Command Command `json:"command" gorm:"foreignKey:CommandID;references:ID;belongsTo:Command"`
}

func CommandDefaults() Command {
	return Command{
		Dir: config.Config.DefaultCommandRunDir,
	}
}

func EmbeddedFileDefaults() EmbeddedFile {
	return EmbeddedFile{}
}

func UserConfigDefaults() UserConfig {
	return UserConfig{
		UsingConsole: config.Config.Console,
	}
}

type RunningCommand interface {
	GetReader() io.Reader
	GetWriter() io.Writer
	Done() <-chan error
	Kill() error
}

type FileParams struct {
	Filename string
	Size     uint64
}

type FileData struct {
	FileId    uint
	CommandId uint
	Bytes     []byte
	Params    FileParams
}
