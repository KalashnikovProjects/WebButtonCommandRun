package errors

import (
	"errors"
)

var ErrNotFound = errors.New("error not found")
var ErrBadName = errors.New("bad object name")
var ErrFileToBig = errors.New("file size too mach")
var ErrEmptyCommand = errors.New("cant run empty command")
