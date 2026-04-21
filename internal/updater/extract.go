package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func extractBinary(tmpDir, archivePath, app string) (string, error) {
	want := app
	if runtime.GOOS == "windows" {
		want += ".exe"
	}

	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		return extractFromZip(tmpDir, archivePath, want)
	case strings.HasSuffix(archivePath, ".tar.gz"):
		return extractFromTarGz(tmpDir, archivePath, want)
	default:
		return "", fmt.Errorf("unsupported archive: %s", archivePath)
	}
}

func extractFromZip(tmpDir, archivePath, wantBase string) (string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = zr.Close() }()

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(f.Name) != wantBase {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer func() { _ = rc.Close() }()

		out := filepath.Join(tmpDir, wantBase)
		if err := writeStreamToFile(out, rc, 0o755); err != nil {
			return "", err
		}
		return out, nil
	}
	return "", errors.New("binary not found in zip")
}

func extractFromTarGz(tmpDir, archivePath, wantBase string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if h.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(h.Name) != wantBase {
			continue
		}
		out := filepath.Join(tmpDir, wantBase)
		if err := writeStreamToFile(out, tr, 0o755); err != nil {
			return "", err
		}
		return out, nil
	}
	return "", errors.New("binary not found in tar.gz")
}

func writeStreamToFile(path string, r io.Reader, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	return f.Close()
}
