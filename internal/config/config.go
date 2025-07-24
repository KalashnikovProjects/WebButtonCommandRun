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

func InitConfigs(rootDir string, envFilename string) error {
	Config = &StructOfConfig{}
	if envFilename != "" {
		if err := godotenv.Load(filepath.Join(rootDir, envFilename)); err != nil {
			return fmt.Errorf("error while loading .env file: %w", err)
		}
	}

	Config.RootDir = rootDir
	Config.UserConfigPath = filepath.Join(Config.RootDir, "data/commands-config.json")
	Config.PORT = os.Getenv("PORT")
	Config.LogLevel = log.Level(map[string]int{"trace": 0, "debug": 1, "info": 2, "warn": 3, "error": 4, "fatal": 5, "panic": 6}[os.Getenv("LogLevel")])
	Config.WebsocketWriteInterval = time.Millisecond * 50
	log.SetLevel(Config.LogLevel)

	Config.Console = DetectDefaultConsole()

	return nil
}
