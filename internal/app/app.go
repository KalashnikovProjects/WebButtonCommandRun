package app

import (
	"fmt"
	consoleCheckerAdapter "github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/console/checker"
	consoleRunnerAdapter "github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/console/runner"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/database"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/filesystem"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/url_opener"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/data"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/runner"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/ui/webserver"
	"github.com/gofiber/fiber/v2/log"
	"time"
)

func Run() {
	err := config.InitConfigs("./../")
	if err != nil {
		log.Fatalw("Error while init configs", "error:", err)
	}
	dbAdapter, err := database.Connect()
	if err != nil {
		log.Fatalw("Error while connecting to storage", "error:", err)
	}
	defer func(db database.DB) {
		err := db.Close()
		if err != nil {
			log.Warnw("Error while closing connection to storage", "error:", err)
		}
	}(dbAdapter)
	fileSystemAdapter, err := filesystem.Connect()
	if err != nil {
		log.Fatalw("Error while connecting to storage", "error:", err)
	}
	dataService := data.NewService(dbAdapter, dbAdapter, fileSystemAdapter)
	consoleChecker := consoleCheckerAdapter.New()
	if err := consoleChecker.CheckAvailability(); err != nil {
		log.Fatalw("Error while checking availability of console", "error:", err)
	}
	runnerAdapter := consoleRunnerAdapter.New()
	runnerService := runner.NewService(runnerAdapter)
	appData := webserver.NewServices(dataService, runnerService)
	app := webserver.CreateApp(*appData)

	if config.Config.OpenURLInBrowser {
		urlOpener := url_opener.New()
		go func() {
			time.Sleep(100 * time.Millisecond)
			err = urlOpener.OpenInBrowser(fmt.Sprintf("http://localhost:%d", config.Config.PORT))
			if err != nil {
				log.Warnw("Error opening url in browser", "error:", err)
			}
		}()
	}

	err = webserver.RunApp(app)
	if err != nil {
		log.Fatalw("Error while running server", "error:", err)
	}
}
