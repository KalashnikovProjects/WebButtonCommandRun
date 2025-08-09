//go:build !windows

package command_runner

import (
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/creack/pty"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"os"
	"os/exec"
	"os/user"
)

type unixCommand struct {
	cmd *exec.Cmd
	pty *os.File
}

func RunCommand(command string, options entities.CommandOptions) (Command, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("error getting user home: %w", err)
	}
	homeDir := usr.HomeDir

	cmd := exec.Command(config.Config.Console, "-c", command)
	cmd.Dir = homeDir
	cmd.Env = append(options.Env, "HOME="+homeDir, "PWD="+homeDir)

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
