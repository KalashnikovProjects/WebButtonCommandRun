package server

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"path/filepath"
)

func RunApp() error {
	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())

	//web := app.Group("/",
	//	cache.New(cache.Config{
	//		Next: func(c *fiber.Ctx) bool {
	//			return c.Query("noCache") == "true" || strings.HasPrefix(c.OriginalURL(), "/api")
	//		},
	//		Expiration:   3 * time.Hour,
	//		CacheControl: true,
	//	}))
	web := app.Group("/")
	web.Get("/", GetIndex)
	web.Static("/static", filepath.Join(config.Config.RootDir, "/web/static"))

	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Post("/commands", PostCommand)
	v1.Get("/commands", GetCommands)
	v1.Get("/commands/:id<min(0)>", GetCommand)
	v1.Patch("/commands/:id<min(0)>", PatchCommand)
	v1.Put("/commands/:id<min(0)>", PutCommand)
	v1.Delete("/commands/:id<min(0)>", DeleteCommand)

	v1.Get("/json-config", GetJsonConfig)
	v1.Post("/json-config", EditJsonConfig)
	v1.Put("/json-config", EditJsonConfig)
	v1.Patch("/json-config", EditJsonConfig)
	v1.Get("/console-using", ConsoleUsing)

	websockets := v1.Group("/ws", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}
		return c.Next()
	})
	websockets.Get("/commands/:id<min(0)>", websocket.New(RunCommandWebsocket))
	return app.Listen(":80")
}
