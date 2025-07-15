package usecases

import (
	"errors"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/json_storage"
	"github.com/aglyzov/go-patch"
)

var ErrCommandIdOutOfRange = json_storage.ErrCommandIdOutOfRange
var ErrCantPatchStruct = errors.New("error cant patch command structure")

func GetUserConfig() (entities.UserConfig, error) {
	return json_storage.GetUserConfig(), nil
}

func SetUserConfig(newConfig entities.UserConfig) error {
	return json_storage.SetUserConfig(newConfig)
}

func AppendCommand(command entities.Command) error {
	return json_storage.AppendCommand(command)
}

func DeleteCommand(commandId uint) error {
	return json_storage.DeleteCommand(commandId)
}

func PatchCommand(commandId uint, newCommand entities.Command) error {
	old, err := json_storage.GetCommand(commandId)
	if err != nil {
		return err
	}
	changed, err := patch.Struct(&old, newCommand)
	if err != nil {
		return ErrCantPatchStruct
	}
	if !changed {
		return nil
	}
	return json_storage.UpdateCommand(commandId, newCommand)
}

func PutCommand(commandId uint, newCommand entities.Command) error {
	return json_storage.UpdateCommand(commandId, newCommand)
}

func GetCommandsList() ([]entities.Command, error) {
	return json_storage.GetCommandsList(), nil
}

func GetCommand(commandId uint) (entities.Command, error) {
	return json_storage.GetCommand(commandId)
}
