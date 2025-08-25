package database

import (
	"fmt"
	"testing"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/testutils"
)

func TestAppendFile(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		file        entities.EmbeddedFile
		expectError bool
	}{
		{
			name: "Append valid file",
			file: entities.EmbeddedFile{
				CommandID: 1,
				Name:      "test.txt",
			},
			expectError: false,
		},
		{
			name: "Append file with empty name",
			file: entities.EmbeddedFile{
				CommandID: 1,
				Name:      "",
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

			err = db.AppendFile(&tc.file)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				// Check that file was added
				files, err := db.GetAllFiles()
				if err != nil {
					t.Fatalf("Cant get files: %v", err)
				}
				if len(files) == 0 {
					t.Fatalf("Expected file to be added, but no files found")
				}
				if files[0].Name != tc.file.Name {
					t.Errorf("Expected name %s, got %s", tc.file.Name, files[0].Name)
				}
			}
		})
	}
}

func TestUpdateFile(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commandID   uint
		fileID      uint
		newFile     entities.EmbeddedFile
		expectError bool
	}{
		{
			name:      "Update existing file",
			commandID: 1,
			fileID:    1,
			newFile: entities.EmbeddedFile{
				Name: "updated.txt",
			},
			expectError: false,
		},
		{
			name:      "Update non-existent file",
			commandID: 1,
			fileID:    999,
			newFile: entities.EmbeddedFile{
				Name: "updated.txt",
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

			// Add a test file first
			if !tc.expectError {
				testFile := entities.EmbeddedFile{
					CommandID: tc.commandID,
					Name:      "test.txt",
				}
				err = db.AppendFile(&testFile)
				if err != nil {
					t.Fatalf("Cant append test file: %v", err)
				}
			}

			err = db.UpdateFile(tc.commandID, tc.fileID, &tc.newFile)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPatchFile(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commandID   uint
		fileID      uint
		newFile     entities.EmbeddedFile
		expectError bool
	}{
		{
			name:      "Patch existing file",
			commandID: 1,
			fileID:    1,
			newFile: entities.EmbeddedFile{
				Name: "patched.txt",
			},
			expectError: false,
		},
		{
			name:      "Patch non-existent file",
			commandID: 1,
			fileID:    999,
			newFile: entities.EmbeddedFile{
				Name: "patched.txt",
			},
			expectError: true,
		},
		{
			name:        "Patch with zero value",
			commandID:   1,
			fileID:      1,
			newFile:     entities.EmbeddedFile{},
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

			// Add a test file first
			if !tc.expectError {
				testFile := entities.EmbeddedFile{
					CommandID: tc.commandID,
					Name:      "test.txt",
				}
				err = db.AppendFile(&testFile)
				if err != nil {
					t.Fatalf("Cant append test file: %v", err)
				}
			}

			err = db.PatchFile(tc.commandID, tc.fileID, &tc.newFile)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDeleteFile(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commandID   uint
		fileID      uint
		expectError bool
	}{
		{
			name:        "Delete existing file",
			commandID:   1,
			fileID:      1,
			expectError: false,
		},
		{
			name:        "Delete non-existent file",
			commandID:   1,
			fileID:      999,
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

			// Add a test file first
			if !tc.expectError {
				testFile := entities.EmbeddedFile{
					CommandID: tc.commandID,
					Name:      "test.txt",
				}
				err = db.AppendFile(&testFile)
				if err != nil {
					t.Fatalf("Cant append test file: %v", err)
				}
			}

			err = db.DeleteFile(tc.commandID, tc.fileID)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetFile(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commandID   uint
		fileID      uint
		expectError bool
	}{
		{
			name:        "Get existing file",
			commandID:   1,
			fileID:      1,
			expectError: false,
		},
		{
			name:        "Get non-existent file",
			commandID:   1,
			fileID:      999,
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

			// Add a test file first
			if !tc.expectError {
				testFile := entities.EmbeddedFile{
					CommandID: tc.commandID,
					Name:      "test.txt",
				}
				err = db.AppendFile(&testFile)
				if err != nil {
					t.Fatalf("Cant append test file: %v", err)
				}
			}

			_, err = db.GetFile(tc.commandID, tc.fileID)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetCommandFiles(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commandID   uint
		expectCount int
		expectError bool
	}{
		{
			name:        "Get files for command with files",
			commandID:   1,
			expectCount: 2,
			expectError: false,
		},
		{
			name:        "Get files for command without files",
			commandID:   2,
			expectCount: 0,
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

			// Add test files for the first test case
			if tc.expectCount > 0 {
				for i := 0; i < tc.expectCount; i++ {
					testFile := entities.EmbeddedFile{
						CommandID: tc.commandID,
						Name:      fmt.Sprintf("file%d.txt", i+1),
					}
					err = db.AppendFile(&testFile)
					if err != nil {
						t.Fatalf("Cant append test file %d: %v", i+1, err)
					}
				}
			}

			files, err := db.GetCommandFiles(tc.commandID)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !tc.expectError {
				if len(files) != tc.expectCount {
					t.Errorf("Expected %d files, got %d", tc.expectCount, len(files))
				}
			}
		})
	}
}

func TestGetAllFiles(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

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

	// Test getting empty files list
	files, err := db.GetAllFiles()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if files == nil {
		t.Fatalf("Expected files slice, got nil")
	}
	if len(files) != 0 {
		t.Errorf("Expected empty files list, got %d files", len(files))
	}

	// Add some test files
	testFiles := []entities.EmbeddedFile{
		{CommandID: 1, Name: "file1.txt"},
		{CommandID: 1, Name: "file2.txt"},
		{CommandID: 2, Name: "file3.txt"},
	}

	for _, file := range testFiles {
		err = db.AppendFile(&file)
		if err != nil {
			t.Fatalf("Cant append test file: %v", err)
		}
	}

	// Test getting all files
	files, err = db.GetAllFiles()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}
}

func TestSetAllFiles(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		files       []entities.EmbeddedFile
		expectError bool
	}{
		{
			name: "Set files list",
			files: []entities.EmbeddedFile{
				{CommandID: 1, Name: "file1.txt"},
				{CommandID: 1, Name: "file2.txt"},
				{CommandID: 2, Name: "file3.txt"},
			},
			expectError: false,
		},
		{
			name:        "Set empty files list",
			files:       []entities.EmbeddedFile{},
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

			err = db.SetAllFiles(tc.files)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				// Check that files were set
				files, err := db.GetAllFiles()
				if err != nil {
					t.Fatalf("Cant get files: %v", err)
				}
				if len(files) != len(tc.files) {
					t.Errorf("Expected %d files, got %d", len(tc.files), len(files))
				}
			}
		})
	}
}

func TestDeleteAllFiles(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

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

	// Add some test files first
	testFiles := []entities.EmbeddedFile{
		{CommandID: 1, Name: "file1.txt"},
		{CommandID: 1, Name: "file2.txt"},
	}

	for _, file := range testFiles {
		err = db.AppendFile(&file)
		if err != nil {
			t.Fatalf("Cant append test file: %v", err)
		}
	}

	// Verify files were added
	files, err := db.GetAllFiles()
	if err != nil {
		t.Fatalf("Cant get files: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("Expected 2 files before deletion, got %d", len(files))
	}

	// Delete all files
	err = db.DeleteAllFiles()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify all files were deleted
	files, err = db.GetAllFiles()
	if err != nil {
		t.Fatalf("Cant get files after deletion: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files after deletion, got %d", len(files))
	}
}

func TestSetCommandFiles(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name        string
		commandID   uint
		files       []entities.EmbeddedFile
		expectError bool
	}{
		{
			name:      "Set files for command",
			commandID: 1,
			files: []entities.EmbeddedFile{
				{Name: "file1.txt"},
				{Name: "file2.txt"},
			},
			expectError: false,
		},
		{
			name:        "Set empty files list for command",
			commandID:   1,
			files:       []entities.EmbeddedFile{},
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

			// Set CommandID for all files
			for i := range tc.files {
				tc.files[i].CommandID = tc.commandID
			}

			err = db.SetCommandFiles(tc.commandID, tc.files)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				// Check that files were set
				files, err := db.GetCommandFiles(tc.commandID)
				if err != nil {
					t.Fatalf("Cant get command files: %v", err)
				}
				if len(files) != len(tc.files) {
					t.Errorf("Expected %d files, got %d", len(tc.files), len(files))
				}
			}
		})
	}
}
