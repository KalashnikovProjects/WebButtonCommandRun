package storage

import (
	"errors"
	"fmt"
	"os"

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
	if _, err := os.Stat(config.Config.DatabasePath); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(config.Config.DatabasePath)
		if err != nil {
			return DB{}, err
		}
		err = f.Close()
		if err != nil {
			return DB{}, err
		}
	}

	db, err := gorm.Open(sqlite.Open(config.Config.DatabasePath), &gorm.Config{})
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

func (db DB) AppendCommand(command entities.Command) error {
	result := db.db.Create(&command)
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) DeleteCommand(id uint) error {
	result := db.db.Delete(&entities.Command{}, id)
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) GetCommands() ([]entities.Command, error) {
	var data []entities.Command
	result := db.db.Order("ID").Find(&data)
	if result.Error != nil {
		return data, fmt.Errorf("error in db operation %w", result.Error)
	}
	return data, nil
}

func (db DB) SetCommands(commands []entities.Command) error {
	err := db.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("1 = 1").Delete(&entities.Command{})
		if result.Error != nil {
			return result.Error
		}
		if len(commands) != 0 {
			result = tx.Create(&commands)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error in db transaction %w", err)
	}
	return err
}

func (db DB) GetCommand(id uint) (entities.Command, error) {
	var data entities.Command
	result := db.db.Take(&data, id)
	if result.Error != nil {
		return data, fmt.Errorf("error in db operation %w", result.Error)
	}
	return data, nil
}

func (db DB) PutCommand(id uint, new entities.Command) error {
	result := db.db.Where("id = ?", id).Select("*").Updates(&new)
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) PatchCommand(id uint, new entities.Command) error {
	result := db.db.Where("id = ?", id).Updates(&new)
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) AppendFile(file entities.EmbeddedFile) error {
	result := db.db.Create(&file)
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) UpdateFile(id uint, new entities.EmbeddedFile) error {
	result := db.db.Where("id = ?", id).Select("*").Updates(&new)
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) PatchFile(id uint, new entities.EmbeddedFile) error {
	result := db.db.Where("id = ?", id).Updates(&new)
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) DeleteFile(id uint) error {
	result := db.db.Delete(&entities.EmbeddedFile{}, id)
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) GetFile(id uint) (entities.EmbeddedFile, error) {
	var data entities.EmbeddedFile
	result := db.db.Where("ID = ?", id).Find(&data)
	if result.Error != nil {
		return data, fmt.Errorf("error in db operation %w", result.Error)
	}
	return data, nil
}

func (db DB) GetCommandFiles(commandId uint) ([]entities.EmbeddedFile, error) {
	var data []entities.EmbeddedFile
	result := db.db.Where("CommandId = ?", commandId).Find(&data)
	if result.Error != nil {
		return data, fmt.Errorf("error in db operation %w", result.Error)
	}
	return data, nil
}

func (db DB) GetAllFiles() ([]entities.EmbeddedFile, error) {
	var data []entities.EmbeddedFile
	result := db.db.Find(&data)
	if result.Error != nil {
		return data, fmt.Errorf("error in db operation %w", result.Error)
	}
	return data, nil
}

func (db DB) SetAllFiles(files []entities.EmbeddedFile) error {
	err := db.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("1 = 1").Delete(&entities.EmbeddedFile{})
		if result.Error != nil {
			return result.Error
		}
		if len(files) != 0 {
			result = tx.Create(&files)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error in db transaction %w", err)
	}
	return err
}

func (db DB) SetCommandFiles(commandId uint, files []entities.EmbeddedFile) error {
	err := db.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("CommandId = ?", commandId).Delete(&entities.EmbeddedFile{})
		if result.Error != nil {
			return result.Error
		}
		if len(files) != 0 {
			result = tx.Create(&files)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error in db transaction %w", err)
	}
	return err
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
