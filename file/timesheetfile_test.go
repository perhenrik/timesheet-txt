package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultFileName(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	got := DefaultFileName()
	want := filepath.Join(homeDir, ".timesheet.txt")

	if got != want {
		t.Fatalf("DefaultFileName() = %q, want %q", got, want)
	}
}
