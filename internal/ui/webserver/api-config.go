package webserver

import (
	"github.com/gofiber/fiber/v2"
)

func (s *Server) getJsonConfig() fiber.Handler {
	return func(c *fiber.Ctx) error {
		conf, err := s.userconfig.GetUserConfig()
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(conf)
	}
}

func (s *Server) editJsonConfig() fiber.Handler {
	return func(c *fiber.Ctx) error {
		conf := s.userconfig.CreateDefaultUserConfig()

		err := c.BodyParser(&conf)
		if err != nil {
			return fiber.ErrBadRequest
		}
		if err := s.userconfig.SetUserConfig(conf); err != nil {
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func (s *Server) consoleUsing() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Send([]byte(s.usingConsole))
	}
}
