package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ErrorNotFound = gorm.ErrRecordNotFound

type DB struct {
	db gorm.DB
}

func Connect() (DB, error) {
	if err := os.MkdirAll(config.Config.DataFolderPath, 0755); err != nil {
		return DB{}, fmt.Errorf("error while creating data folder: %w", err)
	}

	databasePath := filepath.Join(config.Config.DataFolderPath, "data.db")

	if _, err := os.Stat(databasePath); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(databasePath)
		if err != nil {
			return DB{}, err
		}
		err = f.Close()
		if err != nil {
			return DB{}, err
		}
	}

	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return DB{}, fmt.Errorf("error while opening db %w", err)
	}
	err = db.AutoMigrate(&entities.Command{})
	if err != nil {
		return DB{}, fmt.Errorf("cant migrate db %w", err)
	}
	err = db.AutoMigrate(&entities.EmbeddedFile{})
	if err != nil {
		return DB{}, fmt.Errorf("cant migrate db %w", err)
	}
	return DB{db: *db}, nil
}

func (db DB) Close() error {
	dbb, err := db.db.DB()
	if err != nil {
		return fmt.Errorf("error in db operation %w", err)
	}
	err = dbb.Close()
	if err != nil {
		return fmt.Errorf("cant close db %w", err)
	}
	return nil
}
