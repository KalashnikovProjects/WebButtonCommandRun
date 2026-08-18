package database

import (
	"testing"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/testutils"
)

func TestConnect(t *testing.T) {
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

			db, err := Connect(tempDir)
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
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	// First connection should create the database
	db1, err := Connect(tempDir)
	if err != nil {
		t.Fatalf("Cant create first connection: %v", err)
	}

	err = db1.Close()
	if err != nil {
		t.Errorf("Cant close first connection: %v", err)
	}

	// Second connection should connect to existing database
	db2, err := Connect(tempDir)
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
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	db, err := Connect(tempDir)
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}

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
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	db, err := Connect(tempDir)
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
