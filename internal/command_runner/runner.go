package command_runner

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

type ProcessingCommand struct {
	cmd          *exec.Cmd
	OutputReader io.Reader
	InputWriter  io.Writer
}

func RunCommand(ctx context.Context, command string, env []string) (ProcessingCommand, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Env = append(cmd.Environ(), env...)
	output, err := getOutputReader(cmd)
	if err != nil {
		return ProcessingCommand{}, err
	}
	input, err := getInputWriter(cmd)
	if err != nil {
		return ProcessingCommand{}, err
	}
	err = cmd.Start()
	if err != nil {
		return ProcessingCommand{}, fmt.Errorf("error staring command %w", err)
	}
	return ProcessingCommand{cmd: cmd, OutputReader: output, InputWriter: input}, err
}

func getOutputReader(cmd *exec.Cmd) (io.Reader, error) {
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	multiReader := io.MultiReader(outReader, errReader)
	return multiReader, nil
}

func getInputWriter(cmd *exec.Cmd) (io.Writer, error) {
	return cmd.StdinPipe()
}
