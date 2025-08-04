package usecases

import (
	"context"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/acarl005/stripansi"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func normalizeOutput(out string) string {
	return strings.Replace(strings.Replace(strings.Replace(stripansi.Strip(out), "\n\r", "\r", -1), "\r\n", "\r", -1), "\n", "\r", -1)
}

func TestRunCommand_Success(t *testing.T) {
	_ = config.InitConfigs("../..")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	command, err := RunCommand(ctx, "echo hello", entities.CommandOptions{})
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
	_ = config.InitConfigs("../..")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	command, err := RunCommand(ctx, "nonexistentcommand1234", entities.CommandOptions{})
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
	_ = config.InitConfigs("../..")
	ctx, cancel := context.WithCancel(context.Background())
	command, err := RunCommand(ctx, "ping 127.0.0.1", entities.CommandOptions{})
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
	_ = config.InitConfigs("../..")
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

	command, err := RunCommand(ctx, pythonCmd, entities.CommandOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if command.Input == nil || command.Output == nil {
		t.Fatal("command.Input or output channel is nil")
	}
	command.Input <- "1+2\re"
	command.Input <- "xit()\r"
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
		case <-time.After(1 * time.Second):
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
	_ = config.InitConfigs("../..")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tmpFile, err := os.CreateTemp("", "testfile_*.txt")
	if err != nil {
		t.Errorf("cant create tmpFile: %v", err)
	}
	err = tmpFile.Close()
	if err != nil {
		t.Errorf("cant close tmpFile %v", err)
	}

	inputTestText := "Hello from copy 7\x08con!\r"
	outputText := "Hello from copy con!\r"
	var commandText string
	if runtime.GOOS == "windows" {
		commandText = fmt.Sprintf("copy con %s", tmpFile.Name())
	} else {
		commandText = fmt.Sprintf("nano %s", tmpFile.Name())
	}

	command, err := RunCommand(ctx, commandText, entities.CommandOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if command.Input == nil || command.Output == nil {
		t.Fatal("input or output channel is nil")
	}
	command.Input <- inputTestText

	if runtime.GOOS == "windows" {
		command.Input <- "\x1A\r"
	} else {
		command.Input <- "\u0018"
		<-time.After(1 * time.Second)
		command.Input <- "Y\r"
	}

	ok := true
	for ok {
		select {
		case _, ok = <-command.Output:
			if !ok {
				break
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for result")
			return
		}
	}
	tmpFileContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("error while reading temp file with output: %v", err)
	}
	out := normalizeOutput(string(tmpFileContent))
	if out != outputText {
		t.Fatalf("file editing wrong answer. expected: %q, got %q", outputText, out)
	}
}
