package server

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/gofiber/fiber/v2"
	"path/filepath"
)

func (a App) GetIndex(c *fiber.Ctx) error {
	return c.SendFile(filepath.Join(config.Config.RootDir, "/web/templates/index.html"))
}
