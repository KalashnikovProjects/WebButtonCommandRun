package entities

type TerminalOptions struct {
	Cols uint16   `json:"cols"`
	Rows uint16   `json:"rows"`
	Env  []string `json:"-"`
}

type EmbeddedFile struct {
	ID        uint   `json:"id"`
	CommandID uint   `json:"command-id"`
	Name      string `json:"name"`
	DataPath  string `json:"-"`
}

type UserConfig struct {
	UsingConsole string         `json:"using-console"`
	Commands     []Command      `json:"commands"`
	Files        []EmbeddedFile `json:"files"`
}

type Command struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Command string `json:"command"`
}
