package filesystem

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"os"
)

type Adapter struct{}

func Connect() (Adapter, error) {
	err := os.MkdirAll(config.Config.DataFolderPath, 0750)

	if err != nil {
		return Adapter{}, err
	}
	return Adapter{}, nil
}

// TODO
