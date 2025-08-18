package storage

import (
	"fmt"
	"math/rand"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
)

func SetDefaultCommandsNames(commands []entities.Command) {
	for i := 0; i < len(commands); i++ {
		SetDefaultCommandsName(&commands[i])
	}
}

func SetDefaultCommandsName(command *entities.Command) {
	if command.Name == "" {
		command.Name = RandomCommandName()
	}
}

func RandomCommandName() string {
	return fmt.Sprintf("Command %d", rand.Intn(100))
}

func (db DB) AppendCommand(command entities.Command) error {
	SetDefaultCommandsName(&command)
	if err := checkName(command.Name); err != nil {
		return err
	}
	return db.DB.AppendCommand(&command)
}

func (db DB) DeleteCommand(commandId uint) error {
	return db.DB.DeleteCommand(commandId)
}

func (db DB) PatchCommand(commandId uint, newCommand entities.Command) error {
	if newCommand.Name != "" {
		if err := checkName(newCommand.Name); err != nil {
			return err
		}
	}
	return db.DB.PatchCommand(commandId, &newCommand)
}

func (db DB) PutCommand(commandId uint, newCommand entities.Command) error {
	SetDefaultCommandsName(&newCommand)
	if err := checkName(newCommand.Name); err != nil {
		return err
	}
	return db.DB.PutCommand(commandId, &newCommand)
}

func (db DB) GetCommandsList() ([]entities.Command, error) {
	return db.DB.GetCommands()
}

func (db DB) GetCommand(commandId uint) (entities.Command, error) {
	return db.DB.GetCommand(commandId)
}

func (db DB) CommandExists(commandId uint) (bool, error) {
	return db.DB.CommandExists(commandId)
}
