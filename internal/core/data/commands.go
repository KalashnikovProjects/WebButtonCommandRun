package data

import (
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"math/rand"
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

func (s service) AppendCommand(command entities.Command) error {
	SetDefaultCommandsName(&command)
	if err := checkName(command.Name); err != nil {
		return err
	}
	return s.commandsRepo.AppendCommand(&command)
}

func (s service) DeleteCommand(commandId uint) error {
	return s.commandsRepo.DeleteCommand(commandId)
}

func (s service) PatchCommand(commandId uint, newCommand entities.Command) error {
	if newCommand.Name != "" {
		if err := checkName(newCommand.Name); err != nil {
			return err
		}
	}
	return s.commandsRepo.PatchCommand(commandId, &newCommand)
}

func (s service) PutCommand(commandId uint, newCommand entities.Command) error {
	SetDefaultCommandsName(&newCommand)
	if err := checkName(newCommand.Name); err != nil {
		return err
	}
	return s.commandsRepo.PutCommand(commandId, &newCommand)
}

func (s service) GetCommandsList() ([]entities.Command, error) {
	return s.commandsRepo.GetCommands()
}

func (s service) GetCommand(commandId uint) (entities.Command, error) {
	return s.commandsRepo.GetCommand(commandId)
}

func (s service) CommandExists(commandId uint) (bool, error) {
	return s.commandsRepo.CommandExists(commandId)
}
