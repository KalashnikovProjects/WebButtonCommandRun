package json_storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/gofiber/fiber/v2/log"
	"io/fs"
	"os"
)

var ErrCommandIdOutOfRange = errors.New("commands id out of range")

var UserConfig entities.UserConfig

func updateFile() error {
	file, err := os.OpenFile(config.Config.UserConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Warn("Error opening user-config file for updating", err)
		}
	}(file)
	encoder := json.NewEncoder(file)
	err = encoder.Encode(UserConfig)
	if err != nil {
		return fmt.Errorf("cant save user config: %w", err)
	}
	return nil
}

func CreateUserConfigIfInvalid() error {
	data, err := os.ReadFile(config.Config.UserConfigPath)
	UserConfig = entities.UserConfig{
		UsingConsole: config.Config.Console,
		Commands:     make([]entities.Command, 0),
	}
	if err == nil {
		err = json.Unmarshal(data, &UserConfig)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("error while creating user config file: %w", err)
	}
	data, err = os.ReadFile(config.Config.UserConfigPath)
	if err = updateFile(); err != nil {
		return fmt.Errorf("error while creating user config file: %w", err)
	}
	return nil
}

func AppendCommand(command entities.Command) error {
	UserConfig.Commands = append(UserConfig.Commands, command)
	err := updateFile()
	if err != nil {
		return err
	}
	return nil
}

func DeleteCommand(commandId uint) error {
	if commandId >= uint(len(UserConfig.Commands)) {
		return ErrCommandIdOutOfRange
	}
	UserConfig.Commands = append(UserConfig.Commands[:commandId], UserConfig.Commands[commandId+1:]...)
	err := updateFile()
	if err != nil {
		return err
	}
	return nil
}

func GetCommandsList() []entities.Command {
	return UserConfig.Commands
}

func GetCommand(commandId uint) (entities.Command, error) {
	if commandId >= uint(len(UserConfig.Commands)) {
		return entities.Command{}, ErrCommandIdOutOfRange
	}
	return UserConfig.Commands[commandId], nil
}

func UpdateCommand(commandId uint, newCommand entities.Command) error {
	if commandId >= uint(len(UserConfig.Commands)) {
		return ErrCommandIdOutOfRange
	}
	UserConfig.Commands[commandId] = newCommand
	err := updateFile()
	if err != nil {
		return err
	}
	return nil
}

func GetUserConfig() entities.UserConfig {
	return UserConfig
}

func SetUserConfig(newUserConfig entities.UserConfig) error {
	UserConfig = newUserConfig
	err := updateFile()
	if err != nil {
		return err
	}
	return nil
}
