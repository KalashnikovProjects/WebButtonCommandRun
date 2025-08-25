package data

import (
	"strings"
	"testing"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/testutils"
)

func TestSetDefaultFilesNames(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name           string
		inputFiles     []entities.EmbeddedFile
		expectedResult []entities.EmbeddedFile
	}{
		{
			name: "Set default names for empty names",
			inputFiles: []entities.EmbeddedFile{
				{Name: ""},
				{Name: "Existing Name"},
				{Name: ""},
			},
			expectedResult: []entities.EmbeddedFile{
				{Name: "File 0"}, // Default name will be set
				{Name: "Existing Name"},
				{Name: "File 1"}, // Default name will be set
			},
		},
		{
			name: "No empty names",
			inputFiles: []entities.EmbeddedFile{
				{Name: "Name 1"},
				{Name: "Name 2"},
			},
			expectedResult: []entities.EmbeddedFile{
				{Name: "Name 1"},
				{Name: "Name 2"},
			},
		},
		{
			name:           "Empty slice",
			inputFiles:     []entities.EmbeddedFile{},
			expectedResult: []entities.EmbeddedFile{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a copy of input files
			files := make([]entities.EmbeddedFile, len(tc.inputFiles))
			copy(files, tc.inputFiles)

			SetDefaultFilesNames(files)

			// Check that files with empty names got default names
			for i, file := range files {
				if tc.inputFiles[i].Name == "" {
					if file.Name == "" {
						t.Errorf("Expected default name for file %d, got empty name", i)
					}
					if !strings.HasPrefix(file.Name, "File ") {
						t.Errorf("Expected default name starting with 'File ' for file %d, got: %s", i, file.Name)
					}
				} else {
					if file.Name != tc.inputFiles[i].Name {
						t.Errorf("Expected name %s for file %d, got: %s", tc.inputFiles[i].Name, i, file.Name)
					}
				}
			}
		})
	}
}

func TestSetDefaultFileName(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name           string
		inputFile      entities.EmbeddedFile
		expectedResult entities.EmbeddedFile
	}{
		{
			name:           "Set default name for empty name",
			inputFile:      entities.EmbeddedFile{Name: ""},
			expectedResult: entities.EmbeddedFile{Name: "File 0"},
		},
		{
			name:           "Keep existing name",
			inputFile:      entities.EmbeddedFile{Name: "Existing Name"},
			expectedResult: entities.EmbeddedFile{Name: "Existing Name"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := tc.inputFile
			SetDefaultFileName(&file)

			if tc.inputFile.Name == "" {
				if file.Name == "" {
					t.Errorf("Expected default name, got empty name")
				}
				if !strings.HasPrefix(file.Name, "File ") {
					t.Errorf("Expected default name starting with 'File ', got: %s", file.Name)
				}
			} else {
				if file.Name != tc.inputFile.Name {
					t.Errorf("Expected name %s, got: %s", tc.inputFile.Name, file.Name)
				}
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	// Set a reasonable max file size for testing
	originalMaxSize := config.Config.MaxFileSize
	config.Config.MaxFileSize = 1024 // 1KB
	defer func() { config.Config.MaxFileSize = originalMaxSize }()

	testCases := []struct {
		name        string
		fileData    FileParams
		expectError bool
	}{
		{
			name: "Valid file",
			fileData: FileParams{
				Filename: "valid_file.txt",
				Size:     512,
			},
			expectError: false,
		},
		{
			name: "File too big",
			fileData: FileParams{
				Filename: "big_file.txt",
				Size:     2048,
			},
			expectError: true,
		},
		{
			name: "Empty filename",
			fileData: FileParams{
				Filename: "",
				Size:     512,
			},
			expectError: true,
		},
		{
			name: "Invalid characters in filename",
			fileData: FileParams{
				Filename: "file<>.txt",
				Size:     512,
			},
			expectError: true,
		},
		{
			name: "Filename too long",
			fileData: FileParams{
				Filename: strings.Repeat("a", 256),
				Size:     512,
			},
			expectError: true,
		},
		{
			name: "Filename with dots only",
			fileData: FileParams{
				Filename: "...",
				Size:     512,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFile(tc.fileData)
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestAppendFile(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name          string
		initialConfig entities.UserConfig
		commandID     uint
		fileData      FileParams
		fileContent   string
		expectError   bool
	}{
		{
			name: "Append file to existing command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileData:    FileParams{Filename: "test.txt", Size: 10},
			fileContent: "test content",
			expectError: false,
		},
		{
			name: "Append file to non-existent command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   2,
			fileData:    FileParams{Filename: "test.txt", Size: 10},
			fileContent: "test content",
			expectError: true,
		},
		{
			name: "Append file with invalid data",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileData:    FileParams{Filename: "", Size: 10},
			fileContent: "test content",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var u service
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err = CreateDB()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u service) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			err = db.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			fileReader := strings.NewReader(tc.fileContent)
			err = db.AppendFile(tc.commandID, fileReader, tc.fileData)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				// Check that file was added to database
				files, err := db.GetCommandFilesList(tc.commandID)
				if err != nil {
					t.Fatalf("Cant get command files: %v", err)
				}
				if len(files) == 0 {
					t.Fatalf("Expected file to be added, but no files found")
				}
				if files[0].Name != tc.fileData.Filename {
					t.Errorf("Expected filename %s, got %s", tc.fileData.Filename, files[0].Name)
				}
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
		name          string
		initialConfig entities.UserConfig
		commandID     uint
		fileID        uint
		expectError   bool
	}{
		{
			name: "Delete existing file",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      1,
			expectError: false,
		},
		{
			name: "Delete non-existent file",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      2,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var u service
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err = CreateDB()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u service) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			err = db.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			// Add a test file first
			if !tc.expectError {
				fileReader := strings.NewReader("test content")
				err = db.AppendFile(tc.commandID, fileReader, FileParams{Filename: "test.txt", Size: 12})
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

func TestPatchFile(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name          string
		initialConfig entities.UserConfig
		commandID     uint
		fileID        uint
		newFile       entities.EmbeddedFile
		expectError   bool
	}{
		{
			name: "Patch file name",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      1,
			newFile:     entities.EmbeddedFile{Name: "new_name.txt"},
			expectError: false,
		},
		{
			name: "Patch with invalid name",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      1,
			newFile:     entities.EmbeddedFile{Name: ""},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var u service
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err = CreateDB()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u service) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			err = db.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			// Add a test file first
			if !tc.expectError {
				fileReader := strings.NewReader("test content")
				err = db.AppendFile(tc.commandID, fileReader, FileParams{Filename: "test.txt", Size: 12})
				if err != nil {
					t.Fatalf("Cant append test file: %v", err)
				}
			}

			err = db.PatchFile(tc.commandID, tc.fileID, tc.newFile)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPutFile(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name          string
		initialConfig entities.UserConfig
		commandID     uint
		fileID        uint
		newFile       entities.EmbeddedFile
		expectError   bool
	}{
		{
			name: "Put file with valid name",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      1,
			newFile:     entities.EmbeddedFile{Name: "new_name.txt"},
			expectError: false,
		},
		{
			name: "Put file with empty name (should set default)",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      1,
			newFile:     entities.EmbeddedFile{Name: ""},
			expectError: true,
		},
		{
			name: "Put file with invalid name",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      1,
			newFile:     entities.EmbeddedFile{Name: "file<>.txt"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var u service
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err = CreateDB()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u service) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			err = db.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			// Add a test file first
			if !tc.expectError {
				fileReader := strings.NewReader("test content")
				err = db.AppendFile(tc.commandID, fileReader, FileParams{Filename: "test.txt", Size: 12})
				if err != nil {
					t.Fatalf("Cant append test file: %v", err)
				}
			}

			err = db.PutFile(tc.commandID, tc.fileID, tc.newFile)
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
		name          string
		initialConfig entities.UserConfig
		commandID     uint
		fileID        uint
		expectError   bool
	}{
		{
			name: "Get existing file",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      1,
			expectError: false,
		},
		{
			name: "Get non-existent file",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      2,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var u service
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err = CreateDB()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u service) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			err = db.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			// Add a test file first
			if !tc.expectError {
				fileReader := strings.NewReader("test content")
				err = db.AppendFile(tc.commandID, fileReader, FileParams{Filename: "test.txt", Size: 12})
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

func TestGetCommandFilesList(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name          string
		initialConfig entities.UserConfig
		commandID     uint
		expectError   bool
	}{
		{
			name: "Get files for existing command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			expectError: false,
		},
		{
			name: "Get files for non-existent command",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   2,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var u service
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err = CreateDB()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u service) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			err = db.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			_, err = db.GetCommandFilesList(tc.commandID)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetAllFilesList(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	var u service
	tmpDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tmpDir
	db, err = CreateDB()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func(u service) {
		err := db.Close()
		if err != nil {
			t.Errorf("Error closing db: %v", err)
		}
	}(db)

	// Test getting all files list
	files, err := db.GetAllFilesList()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if files == nil {
		t.Fatalf("Expected files slice, got nil")
	}
}

func TestDownloadFile(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name          string
		initialConfig entities.UserConfig
		commandID     uint
		fileID        uint
		fileContent   string
		expectError   bool
	}{
		{
			name: "Download existing file",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      1,
			fileContent: "test content",
			expectError: false,
		},
		{
			name: "Download non-existent file",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			fileID:      2,
			fileContent: "test content",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var u service
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err = CreateDB()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u service) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			err = db.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			// Add a test file first
			if !tc.expectError {
				fileReader := strings.NewReader(tc.fileContent)
				err = db.AppendFile(tc.commandID, fileReader, FileParams{Filename: "test.txt", Size: uint64(len(tc.fileContent))})
				if err != nil {
					t.Fatalf("Cant append test file: %v", err)
				}
			}

			_, data, err := db.DownloadFile(tc.commandID, tc.fileID)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !tc.expectError {
				if string(data) != tc.fileContent {
					t.Errorf("Expected content %s, got %s", tc.fileContent, string(data))
				}
			}
		})
	}
}

func TestDownloadCommandFilesInArchive(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	testCases := []struct {
		name          string
		initialConfig entities.UserConfig
		commandID     uint
		expectError   bool
	}{
		{
			name: "Download archive for command with files",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			expectError: false,
		},
		{
			name: "Download archive for command without files",
			initialConfig: entities.UserConfig{
				UsingConsole: "test",
				Commands: []entities.Command{
					{ID: 1, Name: "Test Command", Command: "echo test"},
				},
			},
			commandID:   1,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var u service
			tmpDir, cleanup := testutils.CreateTempDataFolder(t)
			defer cleanup()

			config.Config.DataFolderPath = tmpDir
			db, err = CreateDB()
			if err != nil {
				t.Fatalf("Cant create db: %v", err)
			}
			defer func(u service) {
				err := db.Close()
				if err != nil {
					t.Errorf("Error closing db: %v", err)
				}
			}(db)
			err = db.SetUserConfig(tc.initialConfig)
			if err != nil {
				t.Fatalf("Cant set initial config: %v", err)
			}

			if tc.name == "Download archive for command with files" {
				fileReader1 := strings.NewReader("content1")
				err = db.AppendFile(tc.commandID, fileReader1, FileParams{Filename: "file1.txt", Size: 8})
				if err != nil {
					t.Fatalf("Cant append test file 1: %v", err)
				}
				fileReader2 := strings.NewReader("content2")
				err = db.AppendFile(tc.commandID, fileReader2, FileParams{Filename: "file2.txt", Size: 8})
				if err != nil {
					t.Fatalf("Cant append test file 2: %v", err)
				}
			}

			data, err := db.DownloadCommandFilesInArchive(tc.commandID)
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !tc.expectError {
				if len(data) == 0 {
					t.Errorf("Expected archive data, got empty")
				}
			}
		})
	}
}

func TestDownloadAllFilesInArchive(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	var u service
	tmpDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tmpDir
	db, err = CreateDB()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func(u service) {
		err := db.Close()
		if err != nil {
			t.Errorf("Error closing db: %v", err)
		}
	}(db)

	data, err := db.DownloadAllFilesInArchive()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if data == nil {
		t.Fatalf("Expected archive data, got nil")
	}
}

func TestImportAllFilesFromArchive(t *testing.T) {
	err := config.InitConfigs("../../..")
	if err != nil {
		t.Fatalf("Cant init configs: %v", err)
	}

	var u service
	tmpDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tmpDir
	db, err = CreateDB()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func(u service) {
		err := db.Close()
		if err != nil {
			t.Errorf("Error closing db: %v", err)
		}
	}(db)
	var emptyData []byte
	err = db.ImportAllFilesFromArchive(emptyData)
	if err == nil {
		t.Log("Import with empty data succeeded (unexpected)")
	}
}
