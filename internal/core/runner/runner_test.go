package runner

import (
	"context"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/console"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/filesystem"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/adapters/storage/database"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/data"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/testutils"
	"github.com/acarl005/stripansi"
	"github.com/gofiber/fiber/v2/log"
)

func normalizeOutput(out string) string {
	return strings.Replace(strings.Replace(strings.Replace(stripansi.Strip(out), "\n\r", "\r", -1), "\r\n", "\r", -1), "\n", "\r", -1)
}

func TestRunCommand_Success(t *testing.T) {
	_ = config.InitConfigs("../../..")
	// create temp data folder
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir

	db, err := database.Connect()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func() { _ = db.Close() }()
	runnerAdapter := console.NewRunner()
	runnerService := NewService(runnerAdapter)
	filesystemAdapter, err := filesystem.Connect()
	if err != nil {
		t.Fatalf("Cant create filesystem connection: %v", err)
	}
	dataService := data.NewService(db, db, filesystemAdapter)

	err = db.SetCommands([]entities.Command{{Name: "Echo", Command: "echo hello", Dir: os.TempDir()}})
	if err != nil {
		t.Fatalf("cant set config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	command, err := runnerService.RunCommand(ctx, dataService, 1, entities.TerminalOptions{Rows: 30, Cols: 120})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if command.Input == nil || command.Output == nil {
		t.Fatal("command.Input or output channel is nil")
	}
	result := ""
	ok := true
	var dataOut string
	for ok {
		select {
		case dataOut, ok = <-command.Output:
			if !ok {
				break
			}
			result += dataOut
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for result")
			return
		}
	}
	out := normalizeOutput(result)
	if out != "hello\r" {
		t.Fatalf("unexpected output: '%q', need 'hello\\r'", out)
	}
}

func TestRunCommand_InvalidCommand(t *testing.T) {
	_ = config.InitConfigs("../../..")
	// create temp data folder
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir
	db, err := database.Connect()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func() { _ = db.Close() }()
	runnerAdapter := console.NewRunner()
	runnerService := NewService(runnerAdapter)
	filesystemAdapter, err := filesystem.Connect()
	if err != nil {
		t.Fatalf("Cant create filesystem connection: %v", err)
	}
	dataService := data.NewService(db, db, filesystemAdapter)
	// seed invalid command
	err = db.SetCommands([]entities.Command{{Name: "Bad", Command: "nonexistentcommand1234", Dir: os.TempDir()}})
	if err != nil {
		t.Fatalf("cant set config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	command, err := runnerService.RunCommand(ctx, dataService, 1, entities.TerminalOptions{Rows: 30, Cols: 120})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := ""
	ok := true
	var dataOut string
	for ok {
		select {
		case dataOut, ok = <-command.Output:
			if !ok {
				break
			}
			result += dataOut
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for result")
			return
		}
	}
	out := normalizeOutput(result)
	if !strings.Contains(out, "nonexistentcommand1234") {
		t.Fatalf("wrong output, need command not finded message, got %s", out)
	}
}

func TestRunCommand_ContextCancel(t *testing.T) {
	_ = config.InitConfigs("../../..")
	ctx, cancel := context.WithCancel(context.Background())
	// create temp Usecases
	// create temp data folder
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir
	db, err := database.Connect()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func() { _ = db.Close() }()
	runnerAdapter := console.NewRunner()
	runnerService := NewService(runnerAdapter)
	filesystemAdapter, err := filesystem.Connect()
	if err != nil {
		t.Fatalf("Cant create filesystem connection: %v", err)
	}
	dataService := data.NewService(db, db, filesystemAdapter)
	// seed long-running command
	err = db.SetCommands([]entities.Command{{Name: "Ping", Command: "ping 127.0.0.1", Dir: os.TempDir()}})
	if err != nil {
		t.Fatalf("cant set config: %v", err)
	}
	command, err := runnerService.RunCommand(ctx, dataService, 1, entities.TerminalOptions{Rows: 30, Cols: 120})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cancel()
	_, ok := <-command.Output
	if !ok {
		return
	}
	select {
	case _, ok := <-command.Output:
		if ok {
			t.Error("output channel not closed or command not finished")
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for output channel to close after context cancel")
	}
}

func TestRunCommand_PythonInteractive(t *testing.T) {
	_ = config.InitConfigs("../../..")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pythonCmd := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err := exec.LookPath("python"); err != nil {
			t.Skip("Python not found (tried python3 and python)")
		} else {
			pythonCmd = "python"
		}
	}

	// create temp data folder
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir
	db, err := database.Connect()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func() { _ = db.Close() }()
	runnerAdapter := console.NewRunner()
	runnerService := NewService(runnerAdapter)
	filesystemAdapter, err := filesystem.Connect()
	if err != nil {
		t.Fatalf("Cant create filesystem connection: %v", err)
	}
	dataService := data.NewService(db, db, filesystemAdapter)
	// seed python command
	err = db.SetCommands([]entities.Command{{Name: "Py", Command: pythonCmd, Dir: os.TempDir()}})
	if err != nil {
		t.Fatalf("cant set config: %v", err)
	}

	command, err := runnerService.RunCommand(ctx, dataService, 1, entities.TerminalOptions{Rows: 30, Cols: 120})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if command.Input == nil || command.Output == nil {
		t.Fatal("command.Input or output channel is nil")
	}
	command.Input <- "1+2\r"
	command.Input <- "exit()\r"
	<-time.After(1 * time.Second)
	result := ""
	ok := true
	var dataOut string
	for ok {
		select {
		case dataOut, ok = <-command.Output:
			if !ok {
				break
			}
			result += dataOut
		case <-time.After(2 * time.Second):
			log.Debug(normalizeOutput(result))
			t.Fatal("timeout waiting for result")
			return
		}
	}
	out := normalizeOutput(result)
	need := ">>> 1+2\r3\r>>> exit()\r"
	if !strings.HasSuffix(out, need) {
		t.Fatalf("unexpected output suffig: '%q', need '%q'", out, need)
	}
}

func TestRunCommand_EditFile(t *testing.T) {
	_ = config.InitConfigs("../../..")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	usr, err := user.Current()
	if err != nil {
		t.Fatalf("error getting current user: %v", err)
	}
	// working directory for command
	workDir, err := os.MkdirTemp(usr.HomeDir, "workdir_*")
	if err != nil {
		t.Fatalf("cant create temp workdir: %v", err)
	}
	defer func() { _ = os.RemoveAll(workDir) }()

	// create temp data folder
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir
	db, err := database.Connect()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func() { _ = db.Close() }()
	runnerAdapter := console.NewRunner()
	runnerService := NewService(runnerAdapter)
	filesystemAdapter, err := filesystem.Connect()
	if err != nil {
		t.Fatalf("Cant create filesystem connection: %v", err)
	}
	dataService := data.NewService(db, db, filesystemAdapter)

	var commandText string
	fileName := "embedded_test.txt"
	if runtime.GOOS == "windows" {
		commandText = fmt.Sprintf("type %s", fileName)
	} else {
		commandText = fmt.Sprintf("cat %s", fileName)
	}
	err = db.SetCommands([]entities.Command{{Name: "WithFile", Command: commandText, Dir: workDir}})
	if err != nil {
		t.Fatalf("cant set config: %v", err)
	}
	// attach embedded file content
	fileContent := []byte("Hello from embedded file\n")
	err = dataService.AppendFile(1, fileContent, entities.FileParams{Filename: fileName, Size: uint64(len(fileContent))})
	if err != nil {
		t.Fatalf("cant append file: %v", err)
	}

	command, err := runnerService.RunCommand(ctx, dataService, 1, entities.TerminalOptions{Rows: 30, Cols: 120})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if command.Input == nil || command.Output == nil {
		t.Fatal("input or output channel is nil")
	}
	// drain output
	result := ""
	ok := true
	var dataOut string
	for ok {
		select {
		case dataOut, ok = <-command.Output:
			if !ok {
				break
			}
			result += dataOut
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for result")
			return
		}
	}
	out := normalizeOutput(result)
	if out != "Hello from embedded file\r" {
		t.Fatalf("unexpected output: %q", out)
	}
	// ensure file was cleaned up from workDir
	if _, err := os.Stat(filepath.Join(workDir, fileName)); !os.IsNotExist(err) {
		t.Fatalf("embedded file was not removed from workDir: %v", err)
	}
}

func TestRunCommand_ExecutionDir(t *testing.T) {
	_ = config.InitConfigs("../../..")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	usr, err := user.Current()
	if err != nil {
		t.Fatalf("error getting current user: %v", err)
	}
	workDir, err := os.MkdirTemp(usr.HomeDir, "workdir_cwd_*")
	if err != nil {
		t.Fatalf("cant create temp workdir: %v", err)
	}
	defer func() { _ = os.RemoveAll(workDir) }()

	// create temp data folder
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir
	db, err := database.Connect()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func() { _ = db.Close() }()
	runnerAdapter := console.NewRunner()
	runnerService := NewService(runnerAdapter)
	filesystemAdapter, err := filesystem.Connect()
	if err != nil {
		t.Fatalf("Cant create filesystem connection: %v", err)
	}
	dataService := data.NewService(db, db, filesystemAdapter)

	var commandText string
	if runtime.GOOS == "windows" {
		commandText = "cd"
	} else {
		commandText = "pwd"
	}
	err = db.SetCommands([]entities.Command{{Name: "CWD", Command: commandText, Dir: workDir}})
	if err != nil {
		t.Fatalf("cant set config: %v", err)
	}

	command, err := runnerService.RunCommand(ctx, dataService, 1, entities.TerminalOptions{Rows: 30, Cols: 120})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// drain output
	result := ""
	ok := true
	var dataOut string
	for ok {
		select {
		case dataOut, ok = <-command.Output:
			if !ok {
				break
			}
			result += dataOut
		case <-time.After(3 * time.Second):
			t.Fatal("timeout waiting for result")
			return
		}
	}
	out := strings.Trim(normalizeOutput(result), "\r")
	if runtime.GOOS == "windows" {
		if !strings.EqualFold(out, workDir) {
			t.Fatalf("unexpected cwd: %q, want %q", out, workDir)
		}
	} else {
		if out != workDir {
			t.Fatalf("unexpected cwd: %q, want %q", out, workDir)
		}
	}
}

func TestRunCommand_WithFile(t *testing.T) {
	_ = config.InitConfigs("../../..")
	// create temp data folder
	tempDir, cleanup := testutils.CreateTempDataFolder(t)
	defer cleanup()

	config.Config.DataFolderPath = tempDir
	db, err := database.Connect()
	if err != nil {
		t.Fatalf("Cant create db: %v", err)
	}
	defer func() { _ = db.Close() }()
	runnerAdapter := console.NewRunner()
	runnerService := NewService(runnerAdapter)
	filesystemAdapter, err := filesystem.Connect()
	if err != nil {
		t.Fatalf("Cant create filesystem connection: %v", err)
	}
	dataService := data.NewService(db, db, filesystemAdapter)
	err = db.SetCommands([]entities.Command{{Name: "Test", Command: "more test-file.txt", Dir: os.TempDir()}})
	if err != nil {
		t.Fatalf("cant set config: %v", err)
	}
	err = dataService.AppendFile(1, []byte("test data"), entities.FileParams{Filename: "test-file.txt", Size: uint64(len([]byte("test data")))})
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	command, err := runnerService.RunCommand(ctx, dataService, 1, entities.TerminalOptions{Rows: 30, Cols: 120})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if command.Input == nil || command.Output == nil {
		t.Fatal("command.Input or output channel is nil")
	}
	result := ""
	ok := true
	var dataOut string
	for ok {
		select {
		case dataOut, ok = <-command.Output:
			if !ok {
				break
			}
			result += dataOut
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for result")
			return
		}
	}
	out := normalizeOutput(result)
	if out != "test data\r" {
		t.Fatalf("unexpected output: %q, need 'test data\\r'", out)
	}
}
