package webserver

import (
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/data"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/runner"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"path/filepath"
	"strings"
	"time"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Services struct {
	Data   data.Service
	Runner runner.Service
}

func NewServices(data data.Service, runner runner.Service) *Services {
	return &Services{
		data,
		runner,
	}
}

func CreateApp(s Services) *fiber.App {
	fiberApp := fiber.New()
	fiberApp.Use(recover.New())
	fiberApp.Use(logger.New())

	web := fiberApp.Group("/",
		cache.New(cache.Config{
			Next: func(c *fiber.Ctx) bool {
				return c.Query("noCache") == "true" || strings.HasPrefix(c.OriginalURL(), "/api")
			},
			Expiration:   3 * time.Hour,
			CacheControl: true,
		}))
	web.Get("/", GetIndex(s))
	web.Static("/static", filepath.Join(config.Config.RootDir, "/web/static"))

	api := fiberApp.Group("/api")
	v1 := api.Group("/v1")

	v1.Post("/commands", PostCommand(s))
	v1.Get("/commands", GetCommands(s))
	v1.Get("/commands/:command_id<min(0)>", GetCommand(s))
	v1.Patch("/commands/:command_id<min(0)>", PatchCommand(s))
	v1.Put("/commands/:command_id<min(0)>", PutCommand(s))
	v1.Delete("/commands/:command_id<min(0)>", DeleteCommand(s))

	v1.Get("/commands/:command_id/files", GetCommandFilesList(s))
	v1.Post("/commands/:command_id<min(0)>/files", PostFiles(s))

	v1.Get("/commands/:command_id<min(0)>/files/:file_id<min(0)>", GetFile(s))
	v1.Put("/commands/:command_id<min(0)>/files/:file_id<min(0)>", PutFile(s))
	v1.Patch("/commands/:command_id<min(0)>/files/:file_id<min(0)>", PatchFile(s))
	v1.Delete("/commands/:command_id<min(0)>/files/:file_id<min(0)>", DeleteFile(s))
	v1.Get("/commands/:command_id<min(0)>/files/:file_id<min(0)>/download", DownloadFile(s))
	v1.Get("/commands/:command_id<min(0)>/files/download", DownloadCommandFiles(s))

	v1.Get("/json-config", GetJsonConfig(s))
	v1.Post("/json-config", EditJsonConfig(s))
	v1.Put("/json-config", EditJsonConfig(s))
	v1.Patch("/json-config", EditJsonConfig(s))

	v1.Get("/files/download", DownloadAllFiles(s))
	v1.Post("/files/upload", ImportFiles(s))

	v1.Get("/console-using", ConsoleUsing(s))

	websockets := v1.Group("/ws", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}
		return c.Next()
	})
	websockets.Get("/commands/:command_id<min(0)>", RunCommandWebsocket(s))
	return fiberApp
}

func RunApp(app *fiber.App) error {
	return app.Listen(fmt.Sprintf(":%d", config.Config.PORT))
}
