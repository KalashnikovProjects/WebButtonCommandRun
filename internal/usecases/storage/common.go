package storage

import (
	"errors"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/storage"
	"os"
	"strings"
)

type DB struct {
	DB *storage.DB
}

var ErrNotFound = storage.ErrorNotFound
var ErrBadName = errors.New("bad file name")

func checkName(name string) error {
	if name == "" {
		return ErrBadName
	}

	invalidChars := []string{
		"<", ">", ":", "\"", "|", "?", "*", "/", "\\",
		"\x00", "\x01", "\x02", "\x03", "\x04", "\x05", "\x06", "\x07",
		"\x08", "\x09", "\x0A", "\x0B", "\x0C", "\x0D", "\x0E", "\x0F",
		"\x10", "\x11", "\x12", "\x13", "\x14", "\x15", "\x16", "\x17",
		"\x18", "\x19", "\x1A", "\x1B", "\x1C", "\x1D", "\x1E", "\x1F",
	}

	result := name
	for _, char := range invalidChars {
		if strings.Contains(result, char) {
			return ErrBadName
		}
	}

	result = strings.TrimSpace(result)
	result = strings.Trim(result, ".")

	if result == "" {
		return ErrBadName
	}

	if len(result) > 255 {
		return ErrBadName
	}

	return nil
}

func CreateDB() (DB, error) {
	err := os.MkdirAll(config.Config.DefaultCommandRunDir, 0755)
	if err != nil {
		return DB{}, err
	}
	db, err := storage.Connect()
	if err != nil {
		return DB{}, err
	}
	return DB{&db}, nil
}

func (db DB) Close() error {
	return db.DB.Close()
}
