package usecases

import (
	"context"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/gofiber/fiber/v2/log"
	"strings"
	"testing"
	"time"
)

func TestRunCommand_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	input, output, err := RunCommand(ctx, "echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input == nil || output == nil {
		t.Fatal("input or output channel is nil")
	}
	select {
	case line, ok := <-output:
		if !ok {
			t.Fatal("output channel closed unexpectedly")
		}
		if line != "hello" && line != "hello\r" { // Windows/Unix
			t.Errorf("unexpected output: %q", line)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for output")
	}
}

func TestRunCommand_InvalidCommand(t *testing.T) {
	_ = config.InitConfigs("../..", "testing.env")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, output, err := RunCommand(ctx, "nonexistentcommand1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outputStr := ""
	for data := range output {
		outputStr += data
	}
	log.Info("Not found command output:", outputStr)
	if !strings.Contains(outputStr, "nonexistentcommand1234") {
		t.Fatalf("wrong output, must command not finded message, got %s", outputStr)
	}
}

func TestRunCommand_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, output, err := RunCommand(ctx, "ping 127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cancel()
	select {
	case _, ok := <-output:
		if ok {
			t.Error("expected output channel to be closed after context cancel")
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for output channel to close after context cancel")
	}
}
