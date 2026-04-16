package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.App.Theme == "" {
		t.Error("expected default theme")
	}
	if cfg.Keymap.Palette == "" {
		t.Error("expected default keymap")
	}
}
