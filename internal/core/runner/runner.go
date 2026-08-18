package runner

import (
	"bufio"
	"context"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Service struct {
	defaultCommandRunDir string
	filesDirPath         string
	runner               Runner
	commands             CommandsRepository
	files                FilesRepository
}

func NewService(defaultCommandRunDir string, filesDirPath string, runner Runner, commandsRepository CommandsRepository, filesRepository FilesRepository) *Service {
	return &Service{
		defaultCommandRunDir: defaultCommandRunDir,
		filesDirPath:         filesDirPath,
		runner:               runner,
		commands:             commandsRepository,
		files:                filesRepository,
	}
}

type deleteCallbackFunction func() error

// prepareFile copy file and return function for delete it
func (s Service) prepareFile(targetDir string, file entities.EmbeddedFile) (deleteCallbackFunction, error) {
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

	sourceFile, err := os.Open(filepath.Join(s.filesDirPath, fmt.Sprintf("%d", file.ID)))
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
func (s Service) RunCommand(ctx context.Context, commandId uint, options entities.TerminalOptions) (*entities.CommandInputOutput, error) {
	commandData, err := s.commands.GetCommand(commandId)
	if err != nil {
		return nil, err
	}
	if commandData.Command == "" {
		return nil, projectErrors.ErrEmptyCommand
	}
	if commandData.Dir == "" {
		options.Dir = s.defaultCommandRunDir
	} else {
		options.Dir = commandData.Dir
	}
	embeddedFiles, err := s.files.GetCommandFiles(commandId)
	if err != nil {
		return nil, err
	}
	var deleteCallbacks []deleteCallbackFunction
	for _, file := range embeddedFiles {
		deleteIt, err := s.prepareFile(options.Dir, file)
		if err != nil {
			return nil, err
		}
		deleteCallbacks = append(deleteCallbacks, deleteIt)
	}
	processingCommand, err := s.runner.RunCommand(commandData.Command, options)
	if err != nil {
		return nil, fmt.Errorf("error in RunCommand function: %w", err)
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
				go func() {
					var err error
					for try := range 3 {
						err = f()
						if err == nil {
							return
						}
						time.Sleep(time.Duration((try+1)*50) * time.Millisecond) // waiting for finishing command executions, for delete its files
					}
					log.Warn(err)
				}()

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
		defer func(processingCommand entities.RunningCommand) {
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
	return &entities.CommandInputOutput{Input: inputChan, Output: outputChan}, nil
}
