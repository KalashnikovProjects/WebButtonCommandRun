package usecases

import (
	"bufio"
	"context"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/command_runner"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/gofiber/fiber/v2/log"
)

type Command struct {
	Input  chan<- string
	Output <-chan string
}

// RunCommand return input chan, output chan and error
func RunCommand(ctx context.Context, commandText string, options entities.CommandOptions) (Command, error) {
	ctx, cancel := context.WithCancel(ctx)
	processingCommand, err := command_runner.RunCommand(commandText, options)
	if err != nil {
		cancel()
		return Command{}, fmt.Errorf("error in RunCommand function: %w", err)
	}

	inputChan := make(chan string)
	outputChan := make(chan string)

	// Output goroutine
	go func() {
		defer close(outputChan)
		defer close(inputChan)
		defer cancel()
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
		defer func(processingCommand command_runner.Command) {
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
