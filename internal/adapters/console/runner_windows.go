//go:build windows

package console

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/iamacarpet/go-winpty"
)

type windowsCommand struct {
	pty *winpty.WinPTY
}

type Runner struct {
}

func NewRunner() *Runner {
	return &Runner{}
}

func (r Runner) RunCommand(command string, options entities.TerminalOptions) (*windowsCommand, error) {
	wp, err := winpty.OpenWithOptions(winpty.Options{
		Dir:         options.Dir,
		DLLPrefix:   filepath.Join(config.Config.RootDir, "pty"),
		Command:     fmt.Sprintf("%s /C %s", config.Config.Console, command),
		Env:         append(append(os.Environ(), "PWD="+options.Dir), options.Env...),
		InitialRows: uint32(options.Rows),
		InitialCols: uint32(options.Cols),
	})
	if err != nil {
		return nil, fmt.Errorf("error failed to get work dir for winpty: %s", err)
	}
	return &windowsCommand{pty: wp}, nil
}

func (c windowsCommand) GetReader() io.Reader {
	return c.pty.StdOut
}

func (c windowsCommand) GetWriter() io.Writer {
	return c.pty.StdIn
}

func (c windowsCommand) Done() <-chan error {
	ch := make(chan error)
	go func() {
		buf := make([]byte, 1)
		for {
			_, err := c.pty.StdOut.Read(buf)
			if err != nil {
				c.pty.Close()
				ch <- err
				return
			}
		}
	}()
	return ch
}

func (c windowsCommand) Kill() error {
	c.pty.Close()
	return nil
}
