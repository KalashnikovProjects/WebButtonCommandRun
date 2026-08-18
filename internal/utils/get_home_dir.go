package utils

import (
	"github.com/gofiber/fiber/v2/log"
	"os/user"
)

func GetHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Error("cant get current user", err)
		return ""
	}
	return usr.HomeDir
}
