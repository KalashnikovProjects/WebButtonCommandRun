package webserver

import (
	"errors"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func PostCommand(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		command := entities.CommandDefaults()
		err := c.BodyParser(&command)
		if err != nil {
			return fiber.ErrBadRequest
		}
		if err := s.Data.AppendCommand(command); err != nil {
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func GetCommands(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commands, err := s.Data.GetCommandsList()
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(commands)
	}
}

func GetCommand(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("command_id")
		if err != nil || id < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		command, err := s.Data.GetCommand(uint(id))
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(command)
	}
}

func PatchCommand(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("command_id")
		if err != nil || id < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		var command entities.Command
		err = c.BodyParser(&command)
		if err != nil {
			return fiber.ErrBadRequest
		}
		err = s.Data.PatchCommand(uint(id), command)
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if errors.Is(err, projectErrors.ErrBadName) {
			return fiber.NewError(fiber.StatusBadRequest, "bad command name")
		} else if err != nil {
			log.Debug(err)
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func PutCommand(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("command_id")
		if err != nil || id < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		command := entities.CommandDefaults()
		err = c.BodyParser(&command)
		if err != nil {
			return fiber.ErrBadRequest
		}
		err = s.Data.PutCommand(uint(id), command)
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if errors.Is(err, projectErrors.ErrBadName) {
			return fiber.NewError(fiber.StatusBadRequest, "bad command name")
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func DeleteCommand(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("command_id")
		if err != nil || id < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		err = s.Data.DeleteCommand(uint(id))
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		}
		return nil
	}
}
