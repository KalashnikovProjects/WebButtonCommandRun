package entities

// TODO: server mode с паролями и заморозкой конфигов

type CommandOptions struct {
	Cols uint16   `json:"cols"`
	Rows uint16   `json:"rows"`
	Env  []string `json:"-"`
}

type UserConfig struct {
	UsingConsole string    `json:"using-console"`
	Commands     []Command `json:"commands"`
}

// Command ID - определяется по индексу в Commands
type Command struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	// TODO возможность заливать alo.sh или alo.cmd файлы вместо команды:
	//  UseFile  bool `json:"use-file"`
	//  FileText bool `json:"file-text"`
}
