package data

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/database"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/filesystem"
	"reflect"
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
		name           string
		initialConfig  entities.UserConfig
		commandToAdd   entities.Command
		expectedConfig entities.UserConfig
		expectError    bool
	}{
		{
			name: "Add command to empty config",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			commandToAdd: entities.Command{
				Name:    "Test Command",
				Command: "echo hello",
			},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{
						ID:      1,
						Name:    "Test Command",
						Command: "echo hello",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Add command to existing config",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{
						Name:    "Existing Command",
						Command: "echo existing",
					},
				},
			},
			commandToAdd: entities.Command{
				Name:    "New Command",
				Command: "echo new",
			},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{
						ID:      1,
						Name:    "Existing Command",
						Command: "echo existing",
					},
					{
						ID:      2,
						Name:    "New Command",
						Command: "echo new",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err := database.Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u database.DB) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			filesystemAdaptor, err := filesystem.Connect()
			if err != nil {
				t.Fatalf("Cant set connect filesystem: %v", err)
			}
			dataService := NewService(db, db, filesystemAdaptor)

			err = dataService.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			err = dataService.AppendCommand(tc.commandToAdd)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resultConfig, err := dataService.GetUserConfig()
			if err != nil {
				t.Fatalf("Cant get result config: %v", err)
			}

			if !tc.expectError {
				if !reflect.DeepEqual(resultConfig, tc.expectedConfig) {
					t.Fatalf("Expected config: %q, got: %q", tc.expectedConfig, resultConfig)
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
		name           string
		initialConfig  entities.UserConfig
		deleteId       uint
		expectedConfig entities.UserConfig
		expectError    bool
	}{
		{
			name: "Delete first command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
					{Name: "Third", Command: "echo third"},
				},
			},
			deleteId: 1,
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 2, Name: "Second", Command: "echo second"},
					{ID: 3, Name: "Third", Command: "echo third"},
				},
			},
			expectError: false,
		},
		{
			name: "Delete middle command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
					{Name: "Third", Command: "echo third"},
				},
			},
			deleteId: 2,
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "First", Command: "echo first"},
					{ID: 3, Name: "Third", Command: "echo third"},
				},
			},
			expectError: false,
		},
		{
			name: "Delete last command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			deleteId: 2,
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "First", Command: "echo first"},
				},
			},
			expectError: false,
		},
		{
			name: "Empty list",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			deleteId: 1,
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands:     []entities.Command{},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err := database.Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u database.DB) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			filesystemAdaptor, err := filesystem.Connect()
			if err != nil {
				t.Fatalf("Cant set connect filesystem: %v", err)
			}
			dataService := NewService(db, db, filesystemAdaptor)

			err = dataService.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			err = dataService.DeleteCommand(tc.deleteId)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resultConfig, err := dataService.GetUserConfig()
			if err != nil {
				t.Fatalf("Cant get result config: %v", err)
			}
			if !tc.expectError {
				if !reflect.DeepEqual(resultConfig, tc.expectedConfig) {
					t.Fatalf("Expected config: %v, got: %v", tc.expectedConfig, resultConfig)
				}
			}
		})
	}
}

func TestGetCommandsList(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name           string
		initialConfig  entities.UserConfig
		expectedResult []entities.Command
	}{
		{
			name: "Get empty list",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			expectedResult: []entities.Command{},
		},
		{
			name: "Get commands list",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			expectedResult: []entities.Command{
				{ID: 1, Name: "First", Command: "echo first"},
				{ID: 2, Name: "Second", Command: "echo second"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err := database.Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u database.DB) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			filesystemAdaptor, err := filesystem.Connect()
			if err != nil {
				t.Fatalf("Cant set connect filesystem: %v", err)
			}
			dataService := NewService(db, db, filesystemAdaptor)

			err = dataService.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}
			result, err := dataService.GetCommandsList()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Fatalf("Expected commands: %v, got: %v", tc.expectedResult, result)
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
		name           string
		initialConfig  entities.UserConfig
		commandId      uint
		expectedResult entities.Command
		expectError    bool
	}{
		{
			name: "Get first command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:      1,
			expectedResult: entities.Command{ID: 1, Name: "First", Command: "echo first"},
			expectError:    false,
		},
		{
			name: "Get second command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:      2,
			expectedResult: entities.Command{ID: 2, Name: "Second", Command: "echo second"},
			expectError:    false,
		},
		{
			name: "Get command out of range",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:      3,
			expectedResult: entities.Command{},
			expectError:    true,
		},
		{
			name: "Get command from empty list",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			commandId:      1,
			expectedResult: entities.Command{},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err := database.Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u database.DB) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			filesystemAdaptor, err := filesystem.Connect()
			if err != nil {
				t.Fatalf("Cant set connect filesystem: %v", err)
			}
			dataService := NewService(db, db, filesystemAdaptor)

			err = dataService.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			result, err := dataService.GetCommand(tc.commandId)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Fatalf("Expected command: %v, got: %v", tc.expectedResult, result)
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
		name           string
		initialConfig  entities.UserConfig
		commandId      uint
		newCommand     entities.Command
		expectedConfig entities.UserConfig
		expectError    bool
	}{
		{
			name: "Update first command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  1,
			newCommand: entities.Command{Name: "Updated First", Command: "echo updated"},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "Updated First", Command: "echo updated"},
					{ID: 2, Name: "Second", Command: "echo second"},
				},
			},
			expectError: false,
		},
		{
			name: "Update second command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  2,
			newCommand: entities.Command{Name: "Updated Second", Command: "echo updated second"},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "First", Command: "echo first"},
					{ID: 2, Name: "Updated Second", Command: "echo updated second"},
				},
			},
			expectError: false,
		},
		{
			name: "Update command out of range",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  3,
			newCommand: entities.Command{Name: "Updated Second", Command: "echo updated second"},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "First", Command: "echo first"},
					{ID: 2, Name: "Second", Command: "echo second"},
				},
			},
			expectError: true,
		},
		{
			name: "Update command in empty list",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			commandId:  1,
			newCommand: entities.Command{Name: "New", Command: "echo new"},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands:     []entities.Command{},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err := database.Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u database.DB) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			filesystemAdaptor, err := filesystem.Connect()
			if err != nil {
				t.Fatalf("Cant set connect filesystem: %v", err)
			}
			dataService := NewService(db, db, filesystemAdaptor)

			err = dataService.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			err = dataService.PutCommand(tc.commandId, tc.newCommand)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resultConfig, err := dataService.GetUserConfig()
			if err != nil {
				t.Fatalf("Cant get result config: %v", err)
			}
			if !tc.expectError {
				if !reflect.DeepEqual(resultConfig, tc.expectedConfig) {
					t.Fatalf("Expected config: %v, got: %v", tc.expectedConfig, resultConfig)
				}
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
		name           string
		initialConfig  entities.UserConfig
		commandId      uint
		newCommand     entities.Command
		expectedConfig entities.UserConfig
		expectError    bool
	}{
		{
			name: "Full patch first command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  1,
			newCommand: entities.Command{Name: "Updated First", Command: "echo updated"},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "Updated First", Command: "echo updated"},
					{ID: 2, Name: "Second", Command: "echo second"},
				},
			},
			expectError: false,
		},
		{
			name: "Patch only name",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  2,
			newCommand: entities.Command{Name: "Updated Second"},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "First", Command: "echo first"},
					{ID: 2, Name: "Updated Second", Command: "echo second"},
				},
			},
			expectError: false,
		},
		{
			name: "Patch only command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  2,
			newCommand: entities.Command{Command: "echo updated second"},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "First", Command: "echo first"},
					{ID: 2, Name: "Second", Command: "echo updated second"},
				},
			},
			expectError: false,
		},
		{
			name: "Update command out of range",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  3,
			newCommand: entities.Command{Name: "Updated Second", Command: "echo updated second"},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "First", Command: "echo first"},
					{ID: 2, Name: "Second", Command: "echo second"},
				},
			},
			expectError: true,
		},
		{
			name: "Update command in empty list",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			commandId:  1,
			newCommand: entities.Command{Name: "New", Command: "echo new"},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands:     []entities.Command{},
			},
			expectError: true,
		},
		{
			name: "No data no changes",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  2,
			newCommand: entities.Command{},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "First", Command: "echo first"},
					{ID: 2, Name: "Second", Command: "echo second"},
				},
			},
			expectError: true,
		},
		{
			name: "Equal data no changes",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  2,
			newCommand: entities.Command{Name: "Second", Command: "echo second"},
			expectedConfig: entities.UserConfig{
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "First", Command: "echo first"},
					{ID: 2, Name: "Second", Command: "echo second"},
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err := database.Connect()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u database.DB) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			filesystemAdaptor, err := filesystem.Connect()
			if err != nil {
				t.Fatalf("Cant set connect filesystem: %v", err)
			}
			dataService := NewService(db, db, filesystemAdaptor)

			err = dataService.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			err = dataService.PatchCommand(tc.commandId, tc.newCommand)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resultConfig, err := dataService.GetUserConfig()
			if err != nil {
				t.Fatalf("Cant get result config: %v", err)
			}
			if !tc.expectError {
				if !reflect.DeepEqual(resultConfig, tc.expectedConfig) {
					t.Fatalf("Expected config: %v, got: %v", tc.expectedConfig, resultConfig)
				}
			}
		})
	}
}
