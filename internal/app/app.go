package app

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/console"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/database"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/filesystem"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/data"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/runner"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/ui/webserver"
	"github.com/gofiber/fiber/v2/log"
)

func Run() {
	err := config.InitConfigs("./")
	if err != nil {
		log.Fatalw("Error while init configs", err)
	}
	dbAdapter, err := database.Connect()
	if err != nil {
		log.Fatalw("Error while connecting to storage", err)
	}
	defer func(db database.DB) {
		err := db.Close()
		if err != nil {
			log.Warnw("Error while closing connection to storage", err)
		}
	}(dbAdapter)
	fileSystemAdapter, err := filesystem.Connect()
	if err != nil {
		log.Fatalw("Error while connecting to storage", err)
	}
	dataService := data.NewService(dbAdapter, dbAdapter, fileSystemAdapter)
	runnerAdapter := console.NewRunner()
	runnerService := runner.NewService(runnerAdapter)
	appData := webserver.NewServices(dataService, runnerService)
	app := webserver.CreateApp(*appData)
	err = webserver.RunApp(app)
	if err != nil {
		log.Fatalw("Error while running server", err)
	}
}
