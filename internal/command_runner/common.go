package command_runner

import (
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
