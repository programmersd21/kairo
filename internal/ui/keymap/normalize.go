package keymap

import (
	"sort"
	"strings"
)

// NormalizeChord normalizes a Bubble Tea-style key chord string so that
// modifier order doesn't matter (e.g. "ctrl+shift+p" == "shift+ctrl+p").
//
// Bubble Tea's Key.String() prefixes Alt first (e.g. "alt+ctrl+p"), which can
// differ from how users write chords in config.
func NormalizeChord(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" || !strings.Contains(s, "+") {
		return s
	}

	parts := strings.Split(s, "+")
	mods := make([]string, 0, len(parts))
	keys := make([]string, 0, 1)

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if isModifier(p) {
			mods = append(mods, p)
		} else {
			keys = append(keys, p)
		}
	}

	if len(keys) == 0 {
		// Something like "ctrl+alt" (no actual key). Return the normalized mods.
		return strings.Join(sortMods(dedupe(mods)), "+")
	}

	key := keys[len(keys)-1]
	mods = sortMods(dedupe(mods))
	if len(mods) == 0 {
		return key
	}
	return strings.Join(append(mods, key), "+")
}

func isModifier(s string) bool {
	switch s {
	case "alt", "ctrl", "shift", "meta":
		return true
	default:
		return false
	}
}

func sortMods(mods []string) []string {
	order := map[string]int{
		"alt":   0,
		"ctrl":  1,
		"shift": 2,
		"meta":  3,
	}
	sort.Slice(mods, func(i, j int) bool {
		oi, okI := order[mods[i]]
		oj, okJ := order[mods[j]]
		if okI && okJ {
			if oi != oj {
				return oi < oj
			}
			return mods[i] < mods[j]
		}
		if okI != okJ {
			return okI
		}
		return mods[i] < mods[j]
	})
	return mods
}

func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
