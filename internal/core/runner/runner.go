package runner

import (
	"bufio"
	"context"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/data"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"os"
	"path/filepath"
)

type RunningCommand interface {
	GetReader() io.Reader
	GetWriter() io.Writer
	Done() <-chan error
	Kill() error
}

type Runner interface {
	RunCommand(command string, options entities.TerminalOptions) (RunningCommand, error)
}

type service struct {
	runner Runner
}

type Service interface {
	RunCommand(ctx context.Context, db data.Service, commandId uint, options entities.TerminalOptions) (Command, error)
}

func NewService(runner Runner) Service {
	return &service{
		runner: runner,
	}
}

type Command struct {
	Input  chan<- string
	Output <-chan string
}

type deleteCallbackFunction func() error

// prepareFile return function for delete file
func prepareFile(targetDir string, file entities.EmbeddedFile) (deleteCallbackFunction, error) {
	targetFileName := filepath.Join(targetDir, file.Name)
	targetFile, err := os.Create(targetFileName)
	if err != nil {
		return nil, err
	}
	defer func(targetFile *os.File) {
		err := targetFile.Close()
		if err != nil {
			log.Warn(err)
		}
	}(targetFile)

	sourceFile, err := os.Open(filepath.Join(config.Config.DataFolderPath, "files", fmt.Sprintf("%d", file.ID)))
	if err != nil {
		return nil, err
	}
	defer func(sourceFile *os.File) {
		err := sourceFile.Close()
		if err != nil {
			log.Warn(err)
		}
	}(sourceFile)

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return nil, err
	}
	return func() error {
		return os.Remove(targetFileName)
	}, nil
}

// RunCommand return input chan, output chan and error
func (s service) RunCommand(ctx context.Context, data data.Service, commandId uint, options entities.TerminalOptions) (Command, error) {
	commandData, err := data.GetCommand(commandId)
	if err != nil {
		return Command{}, err
	}
	if commandData.Command == "" {
		return Command{}, projectErrors.ErrEmptyCommand
	}
	if commandData.Dir == "" {
		options.Dir = config.Config.DefaultCommandRunDir
	} else {
		options.Dir = commandData.Dir
	}
	embeddedFiles, err := data.GetCommandFilesList(commandId)
	if err != nil {
		return Command{}, err
	}
	var deleteCallbacks []deleteCallbackFunction
	for _, file := range embeddedFiles {
		deleteIt, err := prepareFile(options.Dir, file)
		if err != nil {
			return Command{}, err
		}
		deleteCallbacks = append(deleteCallbacks, deleteIt)
	}
	processingCommand, err := s.runner.RunCommand(commandData.Command, options)
	if err != nil {
		return Command{}, fmt.Errorf("error in RunCommand function: %w", err)
	}

	inputChan := make(chan string)
	outputChan := make(chan string)

	ctx, cancel := context.WithCancel(ctx)

	// Output goroutine
	go func() {
		defer close(outputChan)
		defer close(inputChan)
		defer cancel()
		defer func() {
			for _, f := range deleteCallbacks {
				if err := f(); err != nil {
					log.Warn(err)
				}
			}
		}()
		scanner := bufio.NewScanner(processingCommand.GetReader())
		scanner.Split(bufio.ScanRunes)
		for scanner.Scan() {
			select {
			case outputChan <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			log.Debug("Error reading command output", err)
		}
	}()

	// Input goroutine
	go func() {
		defer func(processingCommand RunningCommand) {
			err := processingCommand.Kill()
			if err != nil {
				log.Warn("Error while killing command ", err)
			}
		}(processingCommand)
		for {
			select {
			case input, ok := <-inputChan:
				if !ok {
					return
				}
				_, err := processingCommand.GetWriter().Write([]byte(input))
				if err != nil {
					log.Warn("Error writing input to command", err)
					return
				}
				if flusher, ok := processingCommand.GetWriter().(interface{ Flush() error }); ok {
					if err := flusher.Flush(); err != nil {
						log.Warn("Error flushing input", err)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return Command{Input: inputChan, Output: outputChan}, nil
}
