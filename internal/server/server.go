package server

import (
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/usecases"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"path/filepath"
)

type App struct {
	DB *usecases.DB
}

func CreateApp(app App) *fiber.App {
	fiberApp := fiber.New()
	fiberApp.Use(recover.New())
	fiberApp.Use(logger.New())

	//web := fiberApp.Group("/",
	//	cache.New(cache.Config{
	//		Next: func(c *fiber.Ctx) bool {
	//			return c.Query("noCache") == "true" || strings.HasPrefix(c.OriginalURL(), "/api")
	//		},
	//		Expiration:   3 * time.Hour,
	//		CacheControl: true,
	//	}))
	web := fiberApp.Group("/")
	web.Get("/", app.GetIndex)
	web.Static("/static", filepath.Join(config.Config.RootDir, "/web/static"))

	api := fiberApp.Group("/api")
	v1 := api.Group("/v1")

	v1.Post("/commands", app.PostCommand)
	v1.Get("/commands", app.GetCommands)
	v1.Get("/commands/:id<min(0)>", app.GetCommand)
	v1.Patch("/commands/:id<min(0)>", app.PatchCommand)
	v1.Put("/commands/:id<min(0)>", app.PutCommand)
	v1.Delete("/commands/:id<min(0)>", app.DeleteCommand)

	v1.Get("/json-config", app.GetJsonConfig)
	v1.Post("/json-config", app.EditJsonConfig)
	v1.Put("/json-config", app.EditJsonConfig)
	v1.Patch("/json-config", app.EditJsonConfig)
	v1.Get("/console-using", app.ConsoleUsing)

	websockets := v1.Group("/ws", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}
		return c.Next()
	})
	websockets.Get("/commands/:id<min(0)>", websocket.New(app.RunCommandWebsocket))
	return fiberApp
}

func RunApp(app *fiber.App) error {
	return app.Listen(fmt.Sprintf(":%d", config.Config.PORT))
}
