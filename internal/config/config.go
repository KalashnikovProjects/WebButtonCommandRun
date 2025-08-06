package config

import (
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var portFlag int
var flagsInited bool

type StructOfConfig struct {
	PORT                   int
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
	if !flagsInited {
		flag.IntVar(&portFlag, "port", -1, "port which the server will listen")
		flagsInited = true
	}
	flag.Parse()
	var err error
	Config = &StructOfConfig{}
	envFilename, ok := os.LookupEnv("ENV_FILE")
	if ok {
		if err := godotenv.Load(filepath.Join(rootDir, envFilename)); err != nil {
			return fmt.Errorf("error while loading .env file: %w", err)
		}
	}

	Config.RootDir = rootDir
	Config.UserConfigPath = filepath.Join(Config.RootDir, "data/commands-config.json")

	if portFlag == -1 {
		port, ok := os.LookupEnv("PORT")
		if ok {
			Config.PORT, err = strconv.Atoi(port)
			if err != nil {
				Config.PORT = 8080
			}
		} else {
			Config.PORT = 8080
		}
	} else {
		Config.PORT = portFlag
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
