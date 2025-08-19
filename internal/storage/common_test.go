package storage

import (
	"testing"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/testutils"
)

func TestConnect(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "Connect to database",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tempDir

			db, err := Connect()
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				// Test that we can close the database
				err = db.Close()
				if err != nil {
					t.Errorf("Cant close db: %v", err)
				}
			}
		})
	}
}

func TestConnectToExistingDatabase(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir

	// First connection should create the database
	db1, err := Connect()
	if err != nil {
		t.Fatalf("Cant create first connection: %v", err)
	}

	// Close first connection
	err = db1.Close()
	if err != nil {
		t.Errorf("Cant close first connection: %v", err)
	}

	// Second connection should connect to existing database
	db2, err := Connect()
	if err != nil {
		t.Fatalf("Cant create second connection: %v", err)
	}

	// Close second connection
	err = db2.Close()
	if err != nil {
		t.Errorf("Cant close second connection: %v", err)
	}
}

func TestDBClose(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir

	db, err := Connect()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}

	// Test closing the database
	err = db.Close()
	if err != nil {
		t.Errorf("Cant close db: %v", err)
	}

	// Test closing an already closed database (should not cause issues)
	err = db.Close()
	if err != nil {
		t.Errorf("Cant close already closed db: %v", err)
	}
}

func TestDatabaseMigration(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir

	db, err := Connect()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			t.Errorf("Cant close db: %v", err)
		}
	}()
}
