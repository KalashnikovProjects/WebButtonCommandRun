package errors

import (
	"errors"
)

var WindowsPtyLibUnavailable = errors.New("windows pty library unavailable. Check /pty folder, it must contain dll and exe files")
var ErrNotFound = errors.New("error not found")
var ErrBadName = errors.New("bad object name")
var ErrFileToBig = errors.New("file size too mach")
var ErrEmptyCommand = errors.New("cant run empty command")
