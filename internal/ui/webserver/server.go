package webserver

import (
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	rootDir                string
	port                   int
	usingConsole           string
	maxFileSize            int64
	websocketWriteInterval time.Duration
	commands               Commands
	files                  Files
	userconfig             UserConfig
	runner                 Runner
	fiberApp               *fiber.App
}

func New(rootDir string, port int, usingConsole string, maxFileSize int64, websocketWriteInterval time.Duration, commandsService Commands, filesService Files, userconfigService UserConfig, runner Runner) *Server {
	fiberApp := fiber.New()
	fiberApp.Use(recover.New())
	fiberApp.Use(logger.New())
	s := &Server{
		rootDir,
		port,
		usingConsole,
		maxFileSize,
		websocketWriteInterval,
		commandsService,
		filesService,
		userconfigService,
		runner,
		fiberApp,
	}
	s.bindEndpoints()
	return s
}

func (s *Server) bindEndpoints() {
	web := s.fiberApp.Group("/",
		cache.New(cache.Config{
			Next: func(c *fiber.Ctx) bool {
				return c.Query("noCache") == "true" || strings.HasPrefix(c.OriginalURL(), "/api")
			},
			Expiration:   3 * time.Hour,
			CacheControl: true,
		}))
	web.Get("/", s.getIndex())
	web.Static("/static", filepath.Join(s.rootDir, "/web/static"))

	api := s.fiberApp.Group("/api")
	v1 := api.Group("/v1")

	v1.Post("/commands", s.postCommand())
	v1.Get("/commands", s.getCommands())
	v1.Get("/commands/:command_id<min(0)>", s.getCommand())
	v1.Patch("/commands/:command_id<min(0)>", s.patchCommand())
	v1.Put("/commands/:command_id<min(0)>", s.putCommand())
	v1.Delete("/commands/:command_id<min(0)>", s.deleteCommand())

	v1.Get("/commands/:command_id/files", s.getCommandFilesList())
	v1.Post("/commands/:command_id<min(0)>/files", s.postFiles())

	v1.Get("/commands/:command_id<min(0)>/files/:file_id<min(0)>", s.getFile())
	v1.Put("/commands/:command_id<min(0)>/files/:file_id<min(0)>", s.putFile())
	v1.Patch("/commands/:command_id<min(0)>/files/:file_id<min(0)>", s.patchFile())
	v1.Delete("/commands/:command_id<min(0)>/files/:file_id<min(0)>", s.deleteFile())
	v1.Get("/commands/:command_id<min(0)>/files/:file_id<min(0)>/download", s.downloadFile())
	v1.Get("/commands/:command_id<min(0)>/files/download", s.downloadCommandFiles())

	v1.Get("/json-config", s.getJsonConfig())
	v1.Post("/json-config", s.editJsonConfig())
	v1.Put("/json-config", s.editJsonConfig())
	v1.Patch("/json-config", s.editJsonConfig())

	v1.Get("/files/download", s.downloadAllFiles())
	v1.Post("/files/upload", s.importFiles())

	v1.Get("/console-using", s.consoleUsing())

	websockets := v1.Group("/ws", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}
		return c.Next()
	})
	websockets.Get("/commands/:command_id<min(0)>", s.runCommandWebsocket())
}

func (s *Server) Run() error {
	return s.fiberApp.Listen(fmt.Sprintf(":%d", s.port))
}
