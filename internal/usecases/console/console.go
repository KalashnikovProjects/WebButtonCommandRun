package console

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/console"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/usecases/storage"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"os"
	"path/filepath"
)

type Command struct {
	Input  chan<- string
	Output <-chan string
}

var ErrEmptyCommand = errors.New("cant run empty command")

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
func RunCommand(ctx context.Context, db *storage.DB, commandId uint, options entities.TerminalOptions) (Command, error) {
	commandData, err := db.GetCommand(commandId)
	if err != nil {
		return Command{}, err
	}
	if commandData.Command == "" {
		return Command{}, ErrEmptyCommand
	}
	if commandData.Dir == "" {
		options.Dir = config.Config.DefaultCommandRunDir
	} else {
		options.Dir = commandData.Dir
	}
	embeddedFiles, err := db.GetCommandFilesList(commandId)
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
	processingCommand, err := console.RunCommand(commandData.Command, options)
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
		defer func(processingCommand console.Command) {
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
