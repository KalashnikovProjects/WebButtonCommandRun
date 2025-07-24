package command_runner

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"io"
)

type Command interface {
	GetReader() io.Reader
	GetWriter() io.Writer
	Done() <-chan error
	Kill() error
}

type Options struct {
	Cols uint16
	Rows uint16
	Env  []string
}

func RunCommand(command string, options entities.CommandOptions) (Command, error) {
	if config.Config.Console == "cmd" {
		return RunCommandWindows(command, options)
	} else {
		return RunCommandUnix(command, options)
	}
}
