package testutils

import (
	"os"
	"testing"
)

func CreateTempDataFolder(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "testdata_*")
	if err != nil {
		t.Fatalf("Cant create temp dir: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Cant remove temp dir: %v", err)
		}
	}

	return tempDir, cleanup
}
