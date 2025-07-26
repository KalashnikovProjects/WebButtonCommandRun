package entities

type CommandOptions struct {
	Cols uint16   `json:"cols"`
	Rows uint16   `json:"rows"`
	Env  []string `json:"-"`
}

type UserConfig struct {
	UsingConsole string    `json:"using-console"`
	Commands     []Command `json:"commands"`
}

// Command ID - index of element in UserConfig.Commands
type Command struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}
