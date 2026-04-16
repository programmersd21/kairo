package keymap

import "testing"

func TestNormalizeChord_ModifierOrder(t *testing.T) {
	t.Parallel()

	a := NormalizeChord("ctrl+alt+p")
	b := NormalizeChord("alt+ctrl+p")
	if a != b {
		t.Fatalf("expected chords to normalize equally: %q != %q", a, b)
	}
	if a != "alt+ctrl+p" {
		t.Fatalf("expected bubbletea-style ordering (alt first), got %q", a)
	}
}

func TestNormalizeChord_ShiftTab(t *testing.T) {
	t.Parallel()

	if got := NormalizeChord("shift+tab"); got != "shift+tab" {
		t.Fatalf("expected shift+tab, got %q", got)
	}
	if got := NormalizeChord("tab+shift"); got != "shift+tab" {
		t.Fatalf("expected tab+shift to normalize to shift+tab, got %q", got)
	}
}
