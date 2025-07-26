package config

import (
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type StructOfConfig struct {
	PORT                   string
	RootDir                string
	LogLevel               log.Level
	Console                string // sh or cmd
	UserConfigPath         string
	WebsocketWriteInterval time.Duration
}

var Config *StructOfConfig

func DetectDefaultConsole() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return "sh"
}

func InitConfigs(rootDir string) error {
	Config = &StructOfConfig{}
	envFilename, ok := os.LookupEnv("ENV_FILE")
	if ok {
		if err := godotenv.Load(filepath.Join(rootDir, envFilename)); err != nil {
			return fmt.Errorf("error while loading .env file: %w", err)
		}
	}

	Config.RootDir = rootDir
	Config.UserConfigPath = filepath.Join(Config.RootDir, "data/commands-config.json")
	port, ok := os.LookupEnv("PORT")
	if ok {
		Config.PORT = port
	} else {
		Config.PORT = "80"
	}
	Config.LogLevel = log.Level(map[string]int{"trace": 0, "debug": 1, "info": 2, "warn": 3, "error": 4, "fatal": 5, "panic": 6}[os.Getenv("LogLevel")])
	Config.WebsocketWriteInterval = time.Millisecond * 50
	log.SetLevel(Config.LogLevel)
	console, ok := os.LookupEnv("CONSOLE")
	if ok {
		Config.Console = console
	} else {
		Config.Console = DetectDefaultConsole()
	}

	return nil
}
