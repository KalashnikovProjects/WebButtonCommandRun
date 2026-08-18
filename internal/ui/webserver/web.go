package webserver

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) getIndex() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(s.rootDir, "/web/templates/index.html"))
	}
}
