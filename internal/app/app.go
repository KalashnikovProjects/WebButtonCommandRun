package app

import (
	"fmt"
	consoleCheckerAdapter "github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/console/checker"
	consoleRunnerAdapter "github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/console/runner"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/database"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/filesystem"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/url_opener"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/commands"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/files"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/runner"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/userconfig"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/ui/webserver"
	"github.com/gofiber/fiber/v2/log"
	"path/filepath"
	"time"
)

func Run() {
	err := config.InitConfigs("./../")
	cfg := config.Config
	filesDirPath := filepath.Join(cfg.DataFolderPath, "files")
	ptyDirPath := filepath.Join(cfg.RootDir, "pty")
	if err != nil {
		log.Fatalw("Error while init configs", "error:", err)
	}
	dbAdapter, err := database.Connect(cfg.DataFolderPath)
	if err != nil {
		log.Fatalw("Error while connecting to storage", "error:", err)
	}
	defer func(db database.DB) {
		err := db.Close()
		if err != nil {
			log.Warnw("Error while closing connection to storage", "error:", err)
		}
	}(dbAdapter)
	fileSystemAdapter, err := filesystem.Connect(filesDirPath)
	if err != nil {
		log.Fatalw("Error while connecting to storage", "error:", err)
	}
	consoleChecker := consoleCheckerAdapter.New(ptyDirPath)
	if err := consoleChecker.CheckAvailability(); err != nil {
		log.Fatalw("Error while checking availability of console", "error:", err)
	}
	runnerAdapter := consoleRunnerAdapter.New(ptyDirPath, cfg.Console)

	commandsService := commands.NewService(dbAdapter, cfg.DefaultCommandRunDir)
	filesService := files.NewService(filesDirPath, cfg.MaxFileSize, dbAdapter, dbAdapter, fileSystemAdapter)
	userConfigService := userconfig.NewService(dbAdapter, dbAdapter, fileSystemAdapter, cfg.Console)
	runnerService := runner.NewService(cfg.DefaultCommandRunDir, filesDirPath, runnerAdapter, commandsService, filesService)

	webserverApp := webserver.New(
		cfg.RootDir,
		cfg.PORT,
		cfg.Console,
		cfg.MaxFileSize,
		cfg.WebsocketWriteInterval,
		commandsService,
		filesService,
		userConfigService,
		runnerService,
	)

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

	err = webserverApp.Run()
	if err != nil {
		log.Fatalw("Error while running server", "error:", err)
	}
}
