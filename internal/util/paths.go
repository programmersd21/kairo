package util

import (
	"os"
	"path/filepath"
)

func AppDataDir(appName string) (string, error) {
	d, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "."+appName), nil
}

func AppStateDir(appName string) (string, error) {
	return AppDataDir(appName)
}
