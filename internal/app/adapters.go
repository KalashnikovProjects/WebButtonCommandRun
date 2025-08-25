package app

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/console"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/runner"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
)

type runnerAdapter struct {
	consoleRunner *console.Runner
}

func (r *runnerAdapter) RunCommand(command string, options entities.TerminalOptions) (runner.RunningCommand, error) {
	return r.consoleRunner.RunCommand(command, options)
}
