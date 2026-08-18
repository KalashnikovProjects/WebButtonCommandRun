//go:build windows

package checker

import (
	"errors"
	"fmt"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"os"
	"path/filepath"
)

type Checker struct {
	ptyDir string
}

func New(ptyDir string) *Checker {
	return &Checker{
		ptyDir: ptyDir,
	}
}

func (ch *Checker) CheckAvailability() error {
	if _, err := os.Stat(filepath.Join(ch.ptyDir, "winpty.dll")); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: winpty.dll file unawailable, %w", projectErrors.WindowsPtyLibUnavailable, err)
	}
	if _, err := os.Stat(filepath.Join(ch.ptyDir, "winpty-agent.exe")); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: winpty-agent.exe file unawailable, %w", projectErrors.WindowsPtyLibUnavailable, err)
	}
	return nil
}
