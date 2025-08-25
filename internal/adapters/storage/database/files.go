package database

import (
	"fmt"
	"reflect"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"gorm.io/gorm"
)

func (db DB) AppendFile(file *entities.EmbeddedFile) error {
	result := db.db.Create(file)
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) UpdateFile(commandId, id uint, new *entities.EmbeddedFile) error {
	result := db.db.Where("id = ? and command_id = ?", id, commandId).Select("*").Updates(new)
	if result.RowsAffected == 0 {
		return ErrorNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) PatchFile(commandId, id uint, new *entities.EmbeddedFile) error {
	if reflect.ValueOf(new).IsZero() {
		return nil
	}
	result := db.db.Where("id = ? and command_id = ?", id, commandId).Updates(new)
	if result.RowsAffected == 0 {
		return ErrorNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) DeleteFile(commandId, id uint) error {
	result := db.db.Where("id = ? and command_id = ?", id, commandId).Delete(&entities.EmbeddedFile{})
	if result.RowsAffected == 0 {
		return ErrorNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) GetFile(commandId, id uint) (entities.EmbeddedFile, error) {
	var data entities.EmbeddedFile
	result := db.db.Where("id = ? and command_id = ?", id, commandId).Take(&data)

	if result.Error != nil {
		return data, fmt.Errorf("error in db operation %w", result.Error)
	}
	return data, nil
}

func (db DB) GetCommandFiles(commandId uint) ([]entities.EmbeddedFile, error) {
	var data []entities.EmbeddedFile
	result := db.db.Where("command_id = ?", commandId).Find(&data)
	if result.Error != nil {
		return data, fmt.Errorf("error in db operation %w", result.Error)
	}
	return data, nil
}

func (db DB) GetCommandFilesWithCommandInfo(commandId uint) ([]entities.EmbeddedFileWithCommandInfo, error) {
	var data []entities.EmbeddedFileWithCommandInfo
	result := db.db.Model(&entities.EmbeddedFile{}).Where("command_id = ?", commandId).Preload("Command").Find(&data)
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

func (db DB) GetAllFilesWithCommandInfo() ([]entities.EmbeddedFileWithCommandInfo, error) {
	var data []entities.EmbeddedFileWithCommandInfo
	result := db.db.Model(&entities.EmbeddedFile{}).Preload("Command").Find(&data)
	if result.Error != nil {
		return data, fmt.Errorf("error in db operation %w", result.Error)
	}

	return data, nil
}

func (db DB) SetAllFiles(files []entities.EmbeddedFile) error {
	err := db.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("1=1").Delete(&entities.EmbeddedFile{})
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

func (db DB) DeleteAllFiles() error {
	result := db.db.Where("1=1").Delete(&entities.EmbeddedFile{})
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) SetCommandFiles(commandId uint, files []entities.EmbeddedFile) error {
	err := db.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("command_id = ?", commandId).Delete(&entities.EmbeddedFile{})
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
