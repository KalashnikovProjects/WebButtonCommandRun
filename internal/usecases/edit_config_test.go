package usecases

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"os"
	"reflect"
	"testing"
)

func TestAppendCommand(t *testing.T) {
	err := config.InitConfigs("../..")
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
				UsingConsole: "test",
				Commands: []entities.Command{
					{
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
				UsingConsole: "test",
				Commands: []entities.Command{
					{
						Name:    "Existing Command",
						Command: "echo existing",
					},
					{
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
			tmpFile, err := os.CreateTemp("", "testfile_*.json")
			if err != nil {
				t.Fatalf("Cant create temp file: %v", err)
			}
			defer func(name string) {
				err := tmpFile.Close()
				if err != nil {
					t.Errorf("Cant close temp file: %v", err)
				}
				err = os.Remove(name)
				if err != nil {
					t.Errorf("Cant delete temp file: %v", err)
				}
			}(tmpFile.Name())

			config.Config.UserConfigPath = tmpFile.Name()

			err = SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			err = AppendCommand(tc.commandToAdd)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resultConfig, err := GetUserConfig()
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

func TestDeleteCommand(t *testing.T) {
	err := config.InitConfigs("../..")
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
			deleteId: 0,
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "Second", Command: "echo second"},
					{Name: "Third", Command: "echo third"},
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
			deleteId: 1,
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Third", Command: "echo third"},
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
			deleteId: 1,
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
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
			deleteId: 0,
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "testfile_*.json")
			if err != nil {
				t.Fatalf("Cant create temp file: %v", err)
			}
			defer func(name string) {
				err := tmpFile.Close()
				if err != nil {
					t.Errorf("Cant close temp file: %v", err)
				}
				err = os.Remove(name)
				if err != nil {
					t.Errorf("Cant delete temp file: %v", err)
				}
			}(tmpFile.Name())

			config.Config.UserConfigPath = tmpFile.Name()

			err = SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			err = DeleteCommand(tc.deleteId)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resultConfig, err := GetUserConfig()
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
	err := config.InitConfigs("../..")
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
				{Name: "First", Command: "echo first"},
				{Name: "Second", Command: "echo second"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "testfile_*.json")
			if err != nil {
				t.Fatalf("Cant create temp file: %v", err)
			}
			defer func(name string) {
				err := tmpFile.Close()
				if err != nil {
					t.Errorf("Cant close temp file: %v", err)
				}
				err = os.Remove(name)
				if err != nil {
					t.Errorf("Cant delete temp file: %v", err)
				}
			}(tmpFile.Name())

			config.Config.UserConfigPath = tmpFile.Name()

			err = SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}
			result, err := GetCommandsList()
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
	err := config.InitConfigs("../..")
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
			commandId:      0,
			expectedResult: entities.Command{Name: "First", Command: "echo first"},
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
			commandId:      1,
			expectedResult: entities.Command{Name: "Second", Command: "echo second"},
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
			commandId:      0,
			expectedResult: entities.Command{},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "testfile_*.json")
			if err != nil {
				t.Fatalf("Cant create temp file: %v", err)
			}
			defer func(name string) {
				err := tmpFile.Close()
				if err != nil {
					t.Errorf("Cant close temp file: %v", err)
				}
				err = os.Remove(name)
				if err != nil {
					t.Errorf("Cant delete temp file: %v", err)
				}
			}(tmpFile.Name())

			config.Config.UserConfigPath = tmpFile.Name()

			err = SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			result, err := GetCommand(tc.commandId)
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
	err := config.InitConfigs("../..")
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
			commandId:  0,
			newCommand: entities.Command{Name: "Updated First", Command: "echo updated"},
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "Updated First", Command: "echo updated"},
					{Name: "Second", Command: "echo second"},
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
			commandId:  1,
			newCommand: entities.Command{Name: "Updated Second", Command: "echo updated second"},
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Updated Second", Command: "echo updated second"},
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
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Updated Second", Command: "echo second"},
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
			commandId:  0,
			newCommand: entities.Command{Name: "New", Command: "echo new"},
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "testfile_*.json")
			if err != nil {
				t.Fatalf("Cant create temp file: %v", err)
			}
			defer func(name string) {
				err := tmpFile.Close()
				if err != nil {
					t.Errorf("Cant close temp file: %v", err)
				}
				err = os.Remove(name)
				if err != nil {
					t.Errorf("Cant delete temp file: %v", err)
				}
			}(tmpFile.Name())

			config.Config.UserConfigPath = tmpFile.Name()

			err = SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			err = PutCommand(tc.commandId, tc.newCommand)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resultConfig, err := GetUserConfig()
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
	err := config.InitConfigs("../..")
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
			commandId:  0,
			newCommand: entities.Command{Name: "Updated First", Command: "echo updated"},
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "Updated First", Command: "echo updated"},
					{Name: "Second", Command: "echo second"},
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
			commandId:  1,
			newCommand: entities.Command{Name: "Updated Second"},
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Updated Second", Command: "echo second"},
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
			commandId:  1,
			newCommand: entities.Command{Command: "echo updated second"},
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo updated second"},
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
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Updated Second", Command: "echo second"},
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
			commandId:  0,
			newCommand: entities.Command{Name: "New", Command: "echo new"},
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
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
			commandId:  1,
			newCommand: entities.Command{},
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			expectError: false,
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
			commandId:  1,
			newCommand: entities.Command{Name: "Second", Command: "echo second"},
			expectedConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "testfile_*.json")
			if err != nil {
				t.Fatalf("Cant create temp file: %v", err)
			}
			defer func(name string) {
				err := tmpFile.Close()
				if err != nil {
					t.Errorf("Cant close temp file: %v", err)
				}
				err = os.Remove(name)
				if err != nil {
					t.Errorf("Cant delete temp file: %v", err)
				}
			}(tmpFile.Name())

			config.Config.UserConfigPath = tmpFile.Name()

			err = SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			err = PatchCommand(tc.commandId, tc.newCommand)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resultConfig, err := GetUserConfig()
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

func TestGetUserConfig(t *testing.T) {
	err := config.InitConfigs("../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name           string
		initialConfig  entities.UserConfig
		expectedResult entities.UserConfig
	}{
		{
			name: "Get empty config",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			expectedResult: entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
		},
		{
			name: "Get config with commands",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			expectedResult: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "testfile_*.json")
			if err != nil {
				t.Fatalf("Cant create temp file: %v", err)
			}
			defer func(name string) {
				err := tmpFile.Close()
				if err != nil {
					t.Errorf("Cant close temp file: %v", err)
				}
				err = os.Remove(name)
				if err != nil {
					t.Errorf("Cant delete temp file: %v", err)
				}
			}(tmpFile.Name())

			config.Config.UserConfigPath = tmpFile.Name()

			err = SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			result, err := GetUserConfig()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Fatalf("Expected config: %v, got: %v", tc.expectedResult, result)
			}
		})
	}
}

func TestUpdateUserConfig(t *testing.T) {
	err := config.InitConfigs("../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name           string
		initialConfig  entities.UserConfig
		newConfig      entities.UserConfig
		expectedConfig entities.UserConfig
		expectError    bool
	}{
		{
			name: "Update entire config",
			initialConfig: entities.UserConfig{
				UsingConsole: "old",
				Commands: []entities.Command{
					{Name: "Old", Command: "echo old"},
				},
			},
			newConfig: entities.UserConfig{
				UsingConsole: "new",
				Commands: []entities.Command{
					{Name: "New1", Command: "echo new1"},
					{Name: "New2", Command: "echo new2"},
				},
			},
			expectedConfig: entities.UserConfig{
				UsingConsole: "new",
				Commands: []entities.Command{
					{Name: "New1", Command: "echo new1"},
					{Name: "New2", Command: "echo new2"},
				},
			},
			expectError: false,
		},
		{
			name: "Update to empty config",
			initialConfig: entities.UserConfig{
				UsingConsole: "old",
				Commands: []entities.Command{
					{Name: "Old", Command: "echo old"},
				},
			},
			newConfig: entities.UserConfig{
				UsingConsole: "new",
				Commands:     []entities.Command{},
			},
			expectedConfig: entities.UserConfig{
				UsingConsole: "new",
				Commands:     []entities.Command{},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "testfile_*.json")
			if err != nil {
				t.Fatalf("Cant create temp file: %v", err)
			}
			defer func(name string) {
				err := tmpFile.Close()
				if err != nil {
					t.Errorf("Cant close temp file: %v", err)
				}
				err = os.Remove(name)
				if err != nil {
					t.Errorf("Cant delete temp file: %v", err)
				}
			}(tmpFile.Name())

			config.Config.UserConfigPath = tmpFile.Name()

			err = SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			err = SetUserConfig(tc.newConfig)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resultConfig, err := GetUserConfig()
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
