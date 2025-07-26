package json_storage

import (
	"encoding/json"
	"errors"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"reflect"
	"testing"
)

type structTestCreateUserConfigIfInvalid struct {
	name              string
	fileCreated       bool
	fileContent       string
	randomCommandName bool
	err               error
	resultContent     entities.UserConfig
}

func TestCreateUserConfigIfInvalid(t *testing.T) {
	err := config.InitConfigs("../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}
	currentConsole := config.DetectDefaultConsole()
	defaultResultContent := entities.UserConfig{
		UsingConsole: currentConsole,
		Commands:     make([]entities.Command, 0),
	}
	testData := []structTestCreateUserConfigIfInvalid{
		{
			name:          "File dont exist",
			fileCreated:   false,
			err:           nil,
			resultContent: defaultResultContent,
		},
		{
			name:        "Normal file",
			fileCreated: true,
			fileContent: `{"using-console":"test","commands":[]} `,
			err:         nil,
			resultContent: entities.UserConfig{
				UsingConsole: "test",
				Commands:     make([]entities.Command, 0),
			},
		},
		{
			name:          "Empty file",
			fileCreated:   true,
			fileContent:   "",
			err:           nil,
			resultContent: defaultResultContent,
		},
		{
			name:              "Partial file with redundant fields",
			fileCreated:       true,
			fileContent:       `{"fake-field": 123, "commands":[{"command": "example command", "fake-field-2": "fake"}]}`,
			randomCommandName: true,
			err:               nil,
			resultContent: entities.UserConfig{
				UsingConsole: currentConsole,
				Commands: []entities.Command{
					{
						Name:    "",
						Command: "example command",
					},
				},
			},
		},
		{
			name:          "Broken file",
			fileCreated:   true,
			fileContent:   "aboba",
			err:           nil,
			resultContent: defaultResultContent,
		},
		{
			name:          "Broken file with normal part",
			fileCreated:   true,
			fileContent:   `{"using-console":"test","commands":[]} {"haha": 0}`,
			err:           nil,
			resultContent: defaultResultContent,
		},
		{
			name:          "Another broken file",
			fileCreated:   true,
			fileContent:   `{"using-console":"te"st","commands":[]}`,
			err:           nil,
			resultContent: defaultResultContent,
		},
	}
	for _, testCase := range testData {
		t.Run(testCase.name, func(t *testing.T) {
			var tmpFile *os.File
			config.Config.UserConfigPath = "test123.json"
			var err error
			if testCase.fileCreated {
				tmpFile, err = os.CreateTemp("", "testfile_*.txt")
				if err != nil {
					t.Fatalf("Cant create temp file: %v", err)
				}
				config.Config.UserConfigPath = tmpFile.Name()
				_, err = tmpFile.WriteString(testCase.fileContent)
				if err != nil {
					t.Fatalf("Cant write to file: %v", err)
				}
				err := tmpFile.Close()
				if err != nil {
					t.Errorf("Cant close temp file: %v", err)
				}
			}
			defer func(path string) {
				err := os.Remove(path)
				if err != nil {
					log.Debug(err)
				}
			}(config.Config.UserConfigPath)

			err = CreateUserConfigIfInvalid()
			if !errors.Is(err, testCase.err) {
				if testCase.err == nil {
					t.Fatalf("CreateUserConfigIfInvalid return unexpected error: %v", err)
				} else if err == nil {
					t.Fatalf("CreateUserConfigIfInvalid must return error: %v", testCase.err)
				} else {
					t.Fatalf("CreateUserConfigIfInvalid return wrong error: %v, expected: %v", err, testCase.err)
				}
			}

			content, err := os.ReadFile(config.Config.UserConfigPath)
			if err != nil {
				t.Fatalf("Cant read file after function complete: %v", err)
			}
			jsonData := entities.UserConfig{}
			err = json.Unmarshal(content, &jsonData)
			if err != nil {
				t.Fatalf("Cant unmarshal result: %v", err)
			}
			if testCase.randomCommandName {
				testCase.resultContent.Commands[0].Name = UserConfig.Commands[0].Name
			}
			if !reflect.DeepEqual(jsonData, testCase.resultContent) || !reflect.DeepEqual(jsonData, *UserConfig) {
				t.Fatalf("Wrong answer. Expected: %v, got: %v", testCase.resultContent, jsonData)
			}
		})
	}
}

func TestAppendCommand(t *testing.T) {
	err := config.InitConfigs("../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name           string
		initialConfig  *entities.UserConfig
		commandToAdd   entities.Command
		expectedConfig entities.UserConfig
		expectError    bool
	}{
		{
			name: "Add command to empty config",
			initialConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
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

			UserConfig = tc.initialConfig
			err = updateFile()
			if err != nil {
				t.Fatalf("Cant write initial config: %v", err)
			}

			err = AppendCommand(tc.commandToAdd)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				if !reflect.DeepEqual(*UserConfig, tc.expectedConfig) {
					t.Fatalf("Expected config: %v, got: %v", tc.expectedConfig, UserConfig)
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
		initialConfig  *entities.UserConfig
		deleteId       uint
		expectedConfig entities.UserConfig
		expectError    bool
	}{
		{
			name: "Delete first command",
			initialConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
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

			UserConfig = tc.initialConfig
			err = updateFile()
			if err != nil {
				t.Fatalf("Cant write initial config: %v", err)
			}

			err = DeleteCommand(tc.deleteId)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				if !reflect.DeepEqual(*UserConfig, tc.expectedConfig) {
					t.Fatalf("Expected config: %v, got: %v", tc.expectedConfig, UserConfig)
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
		initialConfig  *entities.UserConfig
		expectedResult []entities.Command
	}{
		{
			name: "Get empty list",
			initialConfig: &entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			expectedResult: []entities.Command{},
		},
		{
			name: "Get commands list",
			initialConfig: &entities.UserConfig{
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

			UserConfig = tc.initialConfig
			err = updateFile()
			if err != nil {
				t.Fatalf("Cant write initial config: %v", err)
			}

			result := GetCommandsList()

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
		initialConfig  *entities.UserConfig
		commandId      uint
		expectedResult entities.Command
		expectError    bool
	}{
		{
			name: "Get first command",
			initialConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
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

			UserConfig = tc.initialConfig
			err = updateFile()
			if err != nil {
				t.Fatalf("Cant write initial config: %v", err)
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

func TestUpdateCommand(t *testing.T) {
	err := config.InitConfigs("../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name           string
		initialConfig  *entities.UserConfig
		commandId      uint
		newCommand     entities.Command
		expectedConfig *entities.UserConfig
		expectError    bool
	}{
		{
			name: "Update first command",
			initialConfig: &entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  0,
			newCommand: entities.Command{Name: "Updated First", Command: "echo updated"},
			expectedConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  1,
			newCommand: entities.Command{Name: "Updated Second", Command: "echo updated second"},
			expectedConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			commandId:  3,
			newCommand: entities.Command{Name: "Updated Second", Command: "echo updated second"},
			expectedConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			commandId:  0,
			newCommand: entities.Command{Name: "New", Command: "echo new"},
			expectedConfig: &entities.UserConfig{
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

			UserConfig = tc.initialConfig
			err = updateFile()
			if err != nil {
				t.Fatalf("Cant write initial config: %v", err)
			}

			err = UpdateCommand(tc.commandId, tc.newCommand)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				if !reflect.DeepEqual(UserConfig, tc.expectedConfig) {
					t.Fatalf("Expected config: %v, got: %v", tc.expectedConfig, UserConfig)
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
		initialConfig  *entities.UserConfig
		expectedResult *entities.UserConfig
	}{
		{
			name: "Get empty config",
			initialConfig: &entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			expectedResult: &entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
		},
		{
			name: "Get config with commands",
			initialConfig: &entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "First", Command: "echo first"},
					{Name: "Second", Command: "echo second"},
				},
			},
			expectedResult: &entities.UserConfig{
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

			UserConfig = tc.initialConfig
			err = updateFile()
			if err != nil {
				t.Fatalf("Cant write initial config: %v", err)
			}

			result := GetUserConfig()

			if !reflect.DeepEqual(result, *tc.expectedResult) {
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
		initialConfig  *entities.UserConfig
		newConfig      entities.UserConfig
		expectedConfig *entities.UserConfig
		expectError    bool
	}{
		{
			name: "Update entire config",
			initialConfig: &entities.UserConfig{
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
			expectedConfig: &entities.UserConfig{
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
			initialConfig: &entities.UserConfig{
				UsingConsole: "old",
				Commands: []entities.Command{
					{Name: "Old", Command: "echo old"},
				},
			},
			newConfig: entities.UserConfig{
				UsingConsole: "new",
				Commands:     []entities.Command{},
			},
			expectedConfig: &entities.UserConfig{
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

			UserConfig = tc.initialConfig
			err = updateFile()
			if err != nil {
				t.Fatalf("Cant write initial config: %v", err)
			}

			err = SetUserConfig(tc.newConfig)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				if !reflect.DeepEqual(UserConfig, tc.expectedConfig) {
					t.Fatalf("Expected config: %v, got: %v", tc.expectedConfig, UserConfig)
				}
			}
		})
	}
}

func TestUpdateFile(t *testing.T) {
	err := config.InitConfigs("../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name           string
		configToWrite  *entities.UserConfig
		expectedResult *entities.UserConfig
		expectError    bool
	}{
		{
			name: "Write empty config",
			configToWrite: &entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			expectedResult: &entities.UserConfig{
				UsingConsole: "test",
				Commands:     []entities.Command{},
			},
			expectError: false,
		},
		{
			name: "Write config with commands",
			configToWrite: &entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "Test1", Command: "echo test1"},
					{Name: "Test2", Command: "echo test2"},
				},
			},
			expectedResult: &entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{Name: "Test1", Command: "echo test1"},
					{Name: "Test2", Command: "echo test2"},
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

			UserConfig = tc.configToWrite

			err = updateFile()
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				content, err := os.ReadFile(config.Config.UserConfigPath)
				if err != nil {
					t.Fatalf("Cant read file after updateFile: %v", err)
				}

				var result *entities.UserConfig
				err = json.Unmarshal(content, &result)
				if err != nil {
					t.Fatalf("Cant unmarshal result: %v", err)
				}

				if !reflect.DeepEqual(result, tc.expectedResult) {
					t.Fatalf("Expected config: %v, got: %v", tc.expectedResult, result)
				}
			}
		})
	}
}
