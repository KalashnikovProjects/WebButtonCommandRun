package utils

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

func SetDefaultFilesNames(files []entities.EmbeddedFile) {
	for i := 0; i < len(files); i++ {
		SetDefaultFileName(&files[i])
	}
}

func SetDefaultFileName(file *entities.EmbeddedFile) {
	if file.Name == "" {
		file.Name = RandomFileName()
	}
}

func RandomFileName() string {
	return fmt.Sprintf("File %d", rand.Intn(100))
}
