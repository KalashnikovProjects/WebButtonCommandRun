package database

import (
	"testing"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/testutils"
)

func TestAppendCommand(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		command     entities.Command
		expectError bool
	}{
		{
			name: "Append valid command",
			command: entities.Command{
				Name:    "Test Command",
				Command: "echo test",
				Dir:     "/tmp",
			},
			expectError: false,
		},
		{
			name: "Append command with empty name",
			command: entities.Command{
				Name:    "",
				Command: "echo test",
				Dir:     "/tmp",
			},
			expectError: false, // GORM allows empty strings
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var db DB
			tempDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tempDir
			db, err = Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("Cant close db: %v", err)
				}
			}()

			err = db.AppendCommand(&tc.command)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				// Check that command was added
				commands, err := db.GetCommands()
				if err != nil {
					t.Fatalf("Cant get commands: %v", err)
				}
				if len(commands) == 0 {
					t.Fatalf("Expected command to be added, but no commands found")
				}
				if commands[0].Name != tc.command.Name {
					t.Errorf("Expected name %s, got %s", tc.command.Name, commands[0].Name)
				}
			}
		})
	}
}

func TestDeleteCommand(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commandID   uint
		expectError bool
	}{
		{
			name:        "Delete existing command",
			commandID:   1,
			expectError: false,
		},
		{
			name:        "Delete non-existent command",
			commandID:   999,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var db DB
			tempDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tempDir
			db, err = Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("Cant close db: %v", err)
				}
			}()

			// Add a test command first
			if !tc.expectError {
				testCommand := entities.Command{
					Name:    "Test Command",
					Command: "echo test",
					Dir:     "/tmp",
				}
				err = db.AppendCommand(&testCommand)
				if err != nil {
					t.Fatalf("Cant append test command: %v", err)
				}
			}

			err = db.DeleteCommand(tc.commandID)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetCommands(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	var db DB
	dbNeedToClose := false
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir
	db, err = Connect()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	dbNeedToClose = true
	defer func() {
		if dbNeedToClose {
			if err := db.Close(); err != nil {
				t.Errorf("Cant close db: %v", err)
			}
		}
	}()

	// Test getting empty commands list
	commands, err := db.GetCommands()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if commands == nil {
		t.Fatalf("Expected commands slice, got nil")
	}
	if len(commands) != 0 {
		t.Errorf("Expected empty commands list, got %d commands", len(commands))
	}

	// Add some test commands
	testCommands := []entities.Command{
		{Name: "Command 1", Command: "echo 1", Dir: "/tmp"},
		{Name: "Command 2", Command: "echo 2", Dir: "/tmp"},
	}

	for _, cmd := range testCommands {
		err = db.AppendCommand(&cmd)
		if err != nil {
			t.Fatalf("Cant append test command: %v", err)
		}
	}

	// Test getting commands list
	commands, err = db.GetCommands()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
}

func TestSetCommands(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commands    []entities.Command
		expectError bool
	}{
		{
			name: "Set commands list",
			commands: []entities.Command{
				{Name: "Command 1", Command: "echo 1", Dir: "/tmp"},
				{Name: "Command 2", Command: "echo 2", Dir: "/tmp"},
			},
			expectError: false,
		},
		{
			name:        "Set empty commands list",
			commands:    []entities.Command{},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var db DB
			tempDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tempDir
			db, err = Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("Cant close db: %v", err)
				}
			}()

			err = db.SetCommands(tc.commands)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				// Check that commands were set
				commands, err := db.GetCommands()
				if err != nil {
					t.Fatalf("Cant get commands: %v", err)
				}
				if len(commands) != len(tc.commands) {
					t.Errorf("Expected %d commands, got %d", len(tc.commands), len(commands))
				}
			}
		})
	}
}

func TestGetCommand(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commandID   uint
		expectError bool
	}{
		{
			name:        "Get existing command",
			commandID:   1,
			expectError: false,
		},
		{
			name:        "Get non-existent command",
			commandID:   999,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var db DB
			tempDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tempDir
			db, err = Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("Cant close db: %v", err)
				}
			}()

			// Add a test command first
			if !tc.expectError {
				testCommand := entities.Command{
					Name:    "Test Command",
					Command: "echo test",
					Dir:     "/tmp",
				}
				err = db.AppendCommand(&testCommand)
				if err != nil {
					t.Fatalf("Cant append test command: %v", err)
				}
			}

			_, err = db.GetCommand(tc.commandID)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPutCommand(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commandID   uint
		newCommand  entities.Command
		expectError bool
	}{
		{
			name:      "Update existing command",
			commandID: 1,
			newCommand: entities.Command{
				Name:    "Updated Command",
				Command: "echo updated",
				Dir:     "/tmp",
			},
			expectError: false,
		},
		{
			name:      "Update non-existent command",
			commandID: 999,
			newCommand: entities.Command{
				Name:    "Updated Command",
				Command: "echo updated",
				Dir:     "/tmp",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var db DB
			tempDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tempDir
			db, err = Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("Cant close db: %v", err)
				}
			}()

			// Add a test command first
			if !tc.expectError {
				testCommand := entities.Command{
					Name:    "Test Command",
					Command: "echo test",
					Dir:     "/tmp",
				}
				err = db.AppendCommand(&testCommand)
				if err != nil {
					t.Fatalf("Cant append test command: %v", err)
				}
			}

			err = db.PutCommand(tc.commandID, &tc.newCommand)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPatchCommand(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commandID   uint
		newCommand  entities.Command
		expectError bool
	}{
		{
			name:      "Patch existing command",
			commandID: 1,
			newCommand: entities.Command{
				Name: "Updated Command",
			},
			expectError: false,
		},
		{
			name:      "Patch non-existent command",
			commandID: 999,
			newCommand: entities.Command{
				Name: "Updated Command",
			},
			expectError: true,
		},
		{
			name:        "Patch with zero value",
			commandID:   1,
			newCommand:  entities.Command{},
			expectError: true, // Should return error for non-existent command
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var db DB
			tempDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tempDir
			db, err = Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("Cant close db: %v", err)
				}
			}()

			// Add a test command first
			if !tc.expectError {
				testCommand := entities.Command{
					Name:    "Test Command",
					Command: "echo test",
					Dir:     "/tmp",
				}
				err = db.AppendCommand(&testCommand)
				if err != nil {
					t.Fatalf("Cant append test command: %v", err)
				}
			}

			err = db.PatchCommand(tc.commandID, &tc.newCommand)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCommandExists(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name         string
		commandID    uint
		expectExists bool
		expectError  bool
	}{
		{
			name:         "Check existing command",
			commandID:    1,
			expectExists: true,
			expectError:  false,
		},
		{
			name:         "Check non-existent command",
			commandID:    999,
			expectExists: false,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var db DB
			tempDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tempDir
			db, err = Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("Cant close db: %v", err)
				}
			}()

			// Add a test command first
			if tc.expectExists {
				testCommand := entities.Command{
					Name:    "Test Command",
					Command: "echo test",
					Dir:     "/tmp",
				}
				err = db.AppendCommand(&testCommand)
				if err != nil {
					t.Fatalf("Cant append test command: %v", err)
				}
			}

			exists, err := db.CommandExists(tc.commandID)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !tc.expectError {
				if exists != tc.expectExists {
					t.Errorf("Expected exists=%v, got %v", tc.expectExists, exists)
				}
			}
		})
	}
}
