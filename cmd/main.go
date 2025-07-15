package main

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/json_storage"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/server"
	"github.com/gofiber/fiber/v2/log"
)

func main() {
	err := config.InitConfigs("./", ".env")
	if err != nil {
		log.Fatalw("Error while init configs", err)
	}
	err = json_storage.CreateUserConfigIfInvalid()
	if err != nil {
		log.Fatalw("Error while initialising user config", err)
	}
	err = server.RunApp()
	if err != nil {
		log.Fatalw("Error while running server", err)
	}
}
