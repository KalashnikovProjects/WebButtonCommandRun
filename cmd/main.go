package main

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/json_storage"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/server"
	"github.com/gofiber/fiber/v2/log"
)

func main() {
	err := config.InitConfigs("./")
	if err != nil {
		log.Fatalw("Error while init configs", err)
	}
	err = json_storage.CreateUserConfigIfInvalid()
	if err != nil {
		log.Fatalw("Error while initialising user config", err)
	}
	app := server.CreateApp()
	err = server.RunApp(app)
	if err != nil {
		log.Fatalw("Error while running server", err)
	}
}
