package search

import "testing"

func TestFuzzyMatchBasic(t *testing.T) {
	m, ok := FuzzyMatch("kairo", "kairo task manager")
	if !ok {
		t.Fatalf("expected match")
	}
	if m.Score <= 0 {
		t.Fatalf("expected positive score, got %d", m.Score)
	}
}

func TestFuzzyMatchPrefersContiguous(t *testing.T) {
	a, okA := FuzzyMatch("abc", "a_b_c")
	b, okB := FuzzyMatch("abc", "abc")
	if !okA || !okB {
		t.Fatalf("expected matches")
	}
	if b.Score <= a.Score {
		t.Fatalf("expected contiguous to score higher: %d vs %d", b.Score, a.Score)
	}
}

func TestFuzzyMatchTypoTolerance(t *testing.T) {
	_, ok := FuzzyMatch("kairoo", "kairo")
	if !ok {
		t.Fatalf("expected typo-tolerant match")
	}
}
