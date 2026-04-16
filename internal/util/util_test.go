package util

import (
	"testing"
)

func TestAppStateDir(t *testing.T) {
	dir, err := AppStateDir("kairo_test")
	if err != nil {
		t.Fatalf("failed to get app dir: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty path")
	}
}
