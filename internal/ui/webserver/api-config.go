package webserver

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/gofiber/fiber/v2"
)

func GetJsonConfig(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		conf, err := s.Data.GetUserConfig()
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(conf)
	}
}

func EditJsonConfig(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		conf := entities.UserConfigDefaults()

		err := c.BodyParser(&conf)
		if err != nil {
			return fiber.ErrBadRequest
		}
		if err := s.Data.SetUserConfig(conf); err != nil {
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func ConsoleUsing(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Send([]byte(config.Config.Console))
	}
}
