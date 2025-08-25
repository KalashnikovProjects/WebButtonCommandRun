package database

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"gorm.io/gorm"
)

func (db DB) AppendCommand(command *entities.Command) error {
	result := db.db.Create(command)
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) DeleteCommand(id uint) error {
	result := db.db.Delete(&entities.Command{}, id)
	if result.RowsAffected == 0 {
		return ErrorNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) GetCommands() ([]entities.Command, error) {
	var data []entities.Command
	result := db.db.Order("ID").Find(&data)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrorNotFound
		} else {
			return nil, fmt.Errorf("error in db operation %w", result.Error)
		}
	}
	return data, nil
}

func (db DB) SetCommands(commands []entities.Command) error {
	err := db.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("1=1").Delete(&entities.Command{})
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
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return entities.Command{}, ErrorNotFound
		} else {
			return entities.Command{}, fmt.Errorf("error in db operation %w", result.Error)
		}
	}
	return data, nil
}

func (db DB) PutCommand(id uint, new *entities.Command) error {
	result := db.db.Where("id = ?", id).Select("*").Updates(&new)
	if result.RowsAffected == 0 {
		return ErrorNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) PatchCommand(id uint, new *entities.Command) error {
	if reflect.ValueOf(new).IsZero() {
		return nil
	}
	result := db.db.Where("id = ?", id).Updates(&new)
	if result.RowsAffected == 0 {
		return ErrorNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("error in db operation %w", result.Error)
	}
	return nil
}

func (db DB) CommandExists(id uint) (bool, error) {
	var count int64
	result := db.db.Model(&entities.Command{}).Where("id = ?", id).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("error in db operation %w", result.Error)
	}
	return count > 0, nil
}
