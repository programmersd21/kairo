package util

import (
	"os"
	"path/filepath"
)

func AppDataDir(appName string) (string, error) {
	if d, err := os.UserConfigDir(); err == nil && d != "" {
		return filepath.Join(d, appName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", appName), nil
}

func AppStateDir(appName string) (string, error) {
	if d, err := os.UserCacheDir(); err == nil && d != "" {
		return filepath.Join(d, appName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", appName), nil
}
