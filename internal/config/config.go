package config

import (
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var portFlag int

type StructOfConfig struct {
	PORT                   int
	RootDir                string
	LogLevel               log.Level
	Console                string // sh or cmd
	DataFolderPath         string
	MaxFileSize            int64 // in bytes, for no restrict <=0
	WebsocketWriteInterval time.Duration
	DefaultCommandRunDir   string
	OpenURLInBrowser       bool
}

var Config *StructOfConfig

func DetectDefaultConsole() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return "sh"
}

func GetHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Error("cant get current user", err)
		return ""
	}
	return usr.HomeDir
}

func flagParse(config *StructOfConfig) {
	if flag.Parsed() {
		return
	}
	flag.IntVar(&portFlag, "port", -1, "port which the server will listen")
	flag.BoolVar(&config.OpenURLInBrowser, "browser", false, "open url of ui in default browser")
	flag.Parse()
}

func InitConfigs(rootDir string) error {
	var err error
	Config = &StructOfConfig{}

	flagParse(Config)
	envFilename, ok := os.LookupEnv("ENV_FILE")
	if ok {
		if err := godotenv.Load(filepath.Join(rootDir, envFilename)); err != nil {
			return fmt.Errorf("error while loading .env file: %w", err)
		}
	}

	Config.RootDir = rootDir
	Config.DataFolderPath = filepath.Join(Config.RootDir, "data")

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
	Config.LogLevel = log.Level(map[string]int{"": 2, "trace": 0, "debug": 1, "info": 2, "warn": 3, "error": 4, "fatal": 5, "panic": 6}[os.Getenv("LOG_LEVEL")])
	Config.MaxFileSize = -1
	Config.WebsocketWriteInterval = time.Millisecond * 50
	Config.DefaultCommandRunDir = GetHomeDir()
	log.SetLevel(Config.LogLevel)
	console, ok := os.LookupEnv("CONSOLE")
	if ok {
		Config.Console = console
	} else {
		Config.Console = DetectDefaultConsole()
	}

	return nil
}
