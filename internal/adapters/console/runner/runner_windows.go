//go:build windows

package runner

import (
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/iamacarpet/go-winpty"
	"io"
	"os"
)

type windowsCommand struct {
	pty *winpty.WinPTY
}

type Runner struct {
	console string // cmd
	ptyDir  string
}

func New(
	ptyDir string,
	console string,
) *Runner {
	return &Runner{
		console: console,
		ptyDir:  ptyDir,
	}
}

func (r Runner) RunCommand(command string, options entities.TerminalOptions) (entities.RunningCommand, error) {
	wp, err := winpty.OpenWithOptions(winpty.Options{
		Dir:         options.Dir,
		DLLPrefix:   r.ptyDir,
		Command:     fmt.Sprintf("%s /C %s", r.console, command),
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
