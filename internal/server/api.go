package server

import (
	"errors"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/usecases"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func PostCommand(c *fiber.Ctx) error {
	var command entities.Command
	err := c.BodyParser(&command)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := usecases.AppendCommand(command); err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func GetCommands(c *fiber.Ctx) error {
	commands, err := usecases.GetCommandsList()
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(commands)
}

func GetCommand(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	command, err := usecases.GetCommand(uint(id))
	if errors.Is(err, usecases.ErrCommandIdOutOfRange) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(command)
}

func PatchCommand(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	var command entities.Command
	err = c.BodyParser(&command)
	if err != nil {
		return fiber.ErrBadRequest
	}
	err = usecases.PatchCommand(uint(id), command)
	if errors.Is(err, usecases.ErrCommandIdOutOfRange) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func PutCommand(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	var command entities.Command
	err = c.BodyParser(&command)
	if err != nil {
		return fiber.ErrBadRequest
	}
	err = usecases.PutCommand(uint(id), command)
	if errors.Is(err, usecases.ErrCommandIdOutOfRange) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func DeleteCommand(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	err = usecases.DeleteCommand(uint(id))
	if errors.Is(err, usecases.ErrCommandIdOutOfRange) {
		return fiber.ErrNotFound
	}
	return nil
}

func GetJsonConfig(c *fiber.Ctx) error {
	conf, err := usecases.GetUserConfig()
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(conf)
}

func EditJsonConfig(c *fiber.Ctx) error {
	conf := entities.UserConfig{
		UsingConsole: config.Config.Console,
		Commands:     []entities.Command{},
	}

	err := c.BodyParser(&conf)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := usecases.SetUserConfig(conf); err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func ConsoleUsing(c *fiber.Ctx) error {
	return c.Send([]byte(config.Config.Console))
}
