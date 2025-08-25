package webserver

import (
	"path/filepath"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/gofiber/fiber/v2"
)

func GetIndex(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(config.Config.RootDir, "/web/templates/index.html"))
	}
}
