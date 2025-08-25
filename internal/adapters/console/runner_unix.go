//go:build !windows

package console

import (
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	coreRunner "github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/runner"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/creack/pty"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"os"
	"os/exec"
)

type unixCommand struct {
	cmd *exec.Cmd
	pty *os.File
}

type Runner struct {
}

func NewRunner() *Runner {
	return &Runner{}
}

func (r runner) RunCommand(command string, options entities.TerminalOptions) (coreRunner.RunningCommand, error) {
	cmd := exec.Command(config.Config.Console, "-c", command)
	cmd.Dir = options.Dir
	cmd.Env = append(options.Env, "PWD="+options.Dir)

	commandPty, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("error starting pty console: %w", err)
	}

	err = pty.Setsize(commandPty, &pty.Winsize{Rows: options.Rows, Cols: options.Cols})
	if err != nil {
		return nil, fmt.Errorf("error updating pty console size: %w", err)
	}

	return unixCommand{cmd: cmd, pty: commandPty}, err
}

func (c unixCommand) GetReader() io.Reader {
	return c.pty
}

func (c unixCommand) GetWriter() io.Writer {
	return c.pty
}

func (c unixCommand) Done() <-chan error {
	ch := make(chan error)
	go func() {
		ch <- c.cmd.Wait()
		log.Debug("Command finished")
	}()
	return ch
}

func (c unixCommand) Kill() error {
	return c.cmd.Process.Kill()
}
