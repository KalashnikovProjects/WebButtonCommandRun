package data

import (
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/database"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/filesystem"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/testutils"
	"reflect"
	"testing"
)

func TestGetUserConfig(t *testing.T) {
	err := config.InitConfigs("../../..")
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
				UsingConsole: config.Config.Console,
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
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 1, Name: "First", Command: "echo first"},
					{ID: 2, Name: "Second", Command: "echo second"},
				},
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

			result, err := dataService.GetUserConfig()
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
	err := config.InitConfigs("../../..")
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
				UsingConsole: config.Config.Console,
				Commands: []entities.Command{
					{ID: 2, Name: "New1", Command: "echo new1"},
					{ID: 3, Name: "New2", Command: "echo new2"},
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
				UsingConsole: config.Config.Console,
				Commands:     []entities.Command{},
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

			err = dataService.SetUserConfig(tc.newConfig)
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
