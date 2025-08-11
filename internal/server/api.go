package server

import (
	"errors"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/usecases"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"strconv"
)

func (a App) PostCommand(c *fiber.Ctx) error {
	var command entities.Command
	err := c.BodyParser(&command)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := a.DB.AppendCommand(command); err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func (a App) GetCommands(c *fiber.Ctx) error {
	commands, err := a.DB.GetCommandsList()
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(commands)
}

func (a App) GetCommand(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	command, err := a.DB.GetCommand(uint(id))
	if errors.Is(err, usecases.ErrCommandNotFound) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(command)
}

func (a App) PatchCommand(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	var command entities.Command
	err = c.BodyParser(&command)
	if err != nil {
		return fiber.ErrBadRequest
	}
	err = a.DB.PatchCommand(uint(id), command)
	if errors.Is(err, usecases.ErrCommandNotFound) {
		return fiber.ErrNotFound
	} else if err != nil {
		log.Debug(err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func (a App) PutCommand(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	var command entities.Command
	err = c.BodyParser(&command)
	if err != nil {
		return fiber.ErrBadRequest
	}
	err = a.DB.PutCommand(uint(id), command)
	if errors.Is(err, usecases.ErrCommandNotFound) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func (a App) DeleteCommand(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	err = a.DB.DeleteCommand(uint(id))
	if errors.Is(err, usecases.ErrCommandNotFound) {
		return fiber.ErrNotFound
	}
	return nil
}

func (a App) GetJsonConfig(c *fiber.Ctx) error {
	conf, err := a.DB.GetUserConfig()
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(conf)
}

func (a App) EditJsonConfig(c *fiber.Ctx) error {
	conf := entities.UserConfig{
		UsingConsole: config.Config.Console,
		Commands:     []entities.Command{},
	}

	err := c.BodyParser(&conf)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := a.DB.SetUserConfig(conf); err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func (a App) ConsoleUsing(c *fiber.Ctx) error {
	return c.Send([]byte(config.Config.Console))
}
