package command_runner

import (
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/iamacarpet/go-winpty"
	"io"
	"os"
	"path/filepath"
)

type windowsCommand struct {
	pty *winpty.WinPTY
}

func RunCommandWindows(command string, options entities.CommandOptions) (Command, error) {
	wp, err := winpty.OpenWithOptions(winpty.Options{
		DLLPrefix:   filepath.Join(config.Config.RootDir, "pty"),
		Command:     fmt.Sprintf("cmd /C %s", command),
		Env:         append(os.Environ(), options.Env...),
		InitialRows: uint32(options.Rows),
		InitialCols: uint32(options.Cols),
	})
	if err != nil {
		return nil, fmt.Errorf("error failed to get work dir for winpty: %s", err)
	}
	return &windowsCommand{pty: wp}, nil
}

func (c *windowsCommand) GetReader() io.Reader {
	return c.pty.StdOut
}

func (c *windowsCommand) GetWriter() io.Writer {
	return c.pty.StdIn
}

func (c *windowsCommand) Done() <-chan error {
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

func (c *windowsCommand) Kill() error {
	c.pty.Close()
	return nil
}
