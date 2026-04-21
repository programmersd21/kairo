package buildinfo

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func VersionTag() string {
	v := strings.TrimSpace(Version)
	if v == "" {
		v = "dev"
	}
	if v != "dev" && !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return v
}

func VersionWithCommit() string {
	v := VersionTag()
	c := strings.TrimSpace(Commit)
	if c == "" || c == "none" {
		return v
	}
	return v + " (" + c + ")"
}

func EffectiveVersion() string {
	if strings.TrimSpace(Version) != "dev" {
		return strings.TrimSpace(Version)
	}
	if data, err := os.ReadFile("VERSION.txt"); err == nil {
		if v := strings.TrimSpace(string(data)); v != "" {
			return v
		}
	}
	if ex, err := os.Executable(); err == nil {
		for _, p := range []string{
			filepath.Join(filepath.Dir(ex), "VERSION.txt"),
			filepath.Join(filepath.Dir(ex), "..", "VERSION.txt"),
		} {
			if data, err := os.ReadFile(p); err == nil {
				if v := strings.TrimSpace(string(data)); v != "" {
					return v
				}
			}
		}
	}
	return "dev"
}
