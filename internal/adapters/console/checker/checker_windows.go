//go:build windows

package checker

import (
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"os"
	"path/filepath"
)

type Checker struct{}

func New() *Checker {
	return &Checker{}
}

func (checker *Checker) CheckAvailability() error {
	if _, err := os.Stat(filepath.Join(config.Config.RootDir, "pty", "winpty.dll")); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: winpty.dll file unawailable, %w", projectErrors.WindowsPtyLibUnavailable, err)
	}
	if _, err := os.Stat(filepath.Join(config.Config.RootDir, "pty", "winpty-agent.exe")); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: winpty-agent.exe file unawailable, %w", projectErrors.WindowsPtyLibUnavailable, err)
	}
	return nil
}
