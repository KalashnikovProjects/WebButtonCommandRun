package server

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/gofiber/fiber/v2"
)

// Import and export files with config
// Maybe change config formate to multipart (json + archive with files)

func (a App) GetJsonConfig(c *fiber.Ctx) error {
	conf, err := a.DB.GetUserConfig()
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(conf)
}

func (a App) EditJsonConfig(c *fiber.Ctx) error {
	conf := entities.UserConfigDefaults()

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
