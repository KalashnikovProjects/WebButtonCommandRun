package webserver

import (
	"errors"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func (s *Server) postCommand() fiber.Handler {
	return func(c *fiber.Ctx) error {
		command := s.commands.DefaultCommand()
		err := c.BodyParser(&command)
		if err != nil {
			return fiber.ErrBadRequest
		}
		if err := s.commands.AppendCommand(command); err != nil {
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func (s *Server) getCommands() fiber.Handler {
	return func(c *fiber.Ctx) error {
		commands, err := s.commands.GetCommandsList()
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(commands)
	}
}

func (s *Server) getCommand() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("command_id")
		if err != nil || id < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		command, err := s.commands.GetCommand(uint(id))
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(command)
	}
}

func (s *Server) patchCommand() fiber.Handler {
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
		err = s.commands.PatchCommand(uint(id), &command)
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

func (s *Server) putCommand() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("command_id")
		if err != nil || id < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		command := s.commands.DefaultCommand()
		err = c.BodyParser(&command)
		if err != nil {
			return fiber.ErrBadRequest
		}
		err = s.commands.PutCommand(uint(id), command)
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

func (s *Server) deleteCommand() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("command_id")
		if err != nil || id < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		err = s.commands.DeleteCommand(uint(id))
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		}
		return nil
	}
}
