package main

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/server"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/usecases/storage"
	"github.com/gofiber/fiber/v2/log"
)

func main() {
	err := config.InitConfigs("./")
	if err != nil {
		log.Fatalw("Error while init configs", err)
	}
	db, err := storage.CreateDB()
	if err != nil {
		log.Fatalw("Error while connecting to storage", err)
	}
	defer func(db storage.DB) {
		err := db.Close()
		if err != nil {
			log.Warnw("Error while closing connection to storage", err)
		}
	}(db)
	appData := server.App{
		DB: &db,
	}
	app := server.CreateApp(appData)
	err = server.RunApp(app)
	if err != nil {
		log.Fatalw("Error while running server", err)
	}
}
