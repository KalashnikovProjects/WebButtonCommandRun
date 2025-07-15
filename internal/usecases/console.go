package usecases

import (
	"bufio"
	"context"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/command_runner"
	"github.com/gofiber/fiber/v2/log"
)

// RunCommand return input chan, output chan and error
func RunCommand(ctx context.Context, commandText string) (chan<- string, <-chan string, error) {
	ctx, cancel := context.WithCancel(ctx)
	processingCommand, err := command_runner.RunCommand(ctx, commandText, make([]string, 0))
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("error in RunCommand function: %w", err)
	}

	inputChan := make(chan string, 100)
	outputChan := make(chan string, 100)

	go func() {
		defer close(outputChan)
		defer close(inputChan)
		defer cancel()
		go func() {
			scanner := bufio.NewScanner(processingCommand.OutputReader)
			for scanner.Scan() {
				select {
				case outputChan <- scanner.Text():
				case <-ctx.Done():
					return
				}
			}
			if err := scanner.Err(); err != nil {
				log.Warn("Error reading command output", err)
			}
		}()

		for {
			select {
			case input, ok := <-inputChan:
				if !ok {
					return
				}
				_, err := processingCommand.InputWriter.Write([]byte(input + "\n"))
				if err != nil {
					log.Warn("Error writing input to command", err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return inputChan, outputChan, nil
}
