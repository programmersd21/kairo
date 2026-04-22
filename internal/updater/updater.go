package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/minio/selfupdate"
)

type Config struct {
	Owner string
	Repo  string
	App   string
}

func DefaultConfig() Config {
	return Config{
		Owner: "programmersd21",
		Repo:  "kairo",
		App:   "kairo",
	}
}

type Release struct {
	TagName string
	Assets  []Asset
}

type Asset struct {
	Name string
	URL  string
}

type CheckResult struct {
	Current string
	Latest  string
	Update  bool
}

func (c Config) Check(ctx context.Context, currentVersion string) (CheckResult, *Release, error) {
	rel, err := c.latestRelease(ctx)
	if err != nil {
		return CheckResult{}, nil, err
	}

	current := strings.TrimSpace(strings.TrimPrefix(currentVersion, "v"))
	latest := strings.TrimSpace(strings.TrimPrefix(rel.TagName, "v"))

	update, err := needsUpdate(current, latest)
	if err != nil {
		return CheckResult{}, nil, err
	}
	return CheckResult{Current: current, Latest: latest, Update: update}, rel, nil
}

type UpdateOptions struct {
	CurrentVersion string
	Stdout         io.Writer
	Stderr         io.Writer
	CheckOnly      bool
}

func (c Config) Update(ctx context.Context, opts UpdateOptions) error {
	out := opts.Stdout
	if out == nil {
		out = os.Stdout
	}
	errOut := opts.Stderr
	if errOut == nil {
		errOut = os.Stderr
	}

	res, rel, err := c.Check(ctx, opts.CurrentVersion)
	if err != nil {
		return err
	}

	if !res.Update {
		_, _ = fmt.Fprintf(out, "✓ Already up to date (%s)\n", displayVersion(res.Current))
		return nil
	}

	_, _ = fmt.Fprintf(out, "Update available: %s → %s\n", displayVersion(res.Current), displayVersion(res.Latest))
	if opts.CheckOnly {
		return nil
	}

	archiveName := expectedArchiveName(c.App)
	archive, ok := findAsset(rel.Assets, archiveName)
	if !ok {
		return fmt.Errorf("release asset not found: %s", archiveName)
	}
	checksums, ok := findAsset(rel.Assets, "checksums.txt")
	if !ok {
		return errors.New("release asset not found: checksums.txt")
	}

	tmpDir, err := os.MkdirTemp("", c.App+"-update-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	client := &http.Client{Timeout: 5 * time.Minute}

	checksumPath := filepath.Join(tmpDir, "checksums.txt")
	if err := downloadToFile(ctx, client, checksums.URL, checksumPath); err != nil {
		return fmt.Errorf("download checksums: %w", err)
	}
	wantSums, err := parseChecksums(checksumPath)
	if err != nil {
		return fmt.Errorf("parse checksums: %w", err)
	}
	wantSum, ok := wantSums[archiveName]
	if !ok {
		return fmt.Errorf("checksum for %s not found in checksums.txt", archiveName)
	}

	archivePath := filepath.Join(tmpDir, archiveName)
	if err := downloadToFile(ctx, client, archive.URL, archivePath); err != nil {
		return fmt.Errorf("download archive: %w", err)
	}
	gotSum, err := sha256FileHex(archivePath)
	if err != nil {
		return fmt.Errorf("hash archive: %w", err)
	}
	if !strings.EqualFold(gotSum, wantSum) {
		return fmt.Errorf("checksum mismatch for %s (want %s got %s)", archiveName, wantSum, gotSum)
	}

	newBinPath, err := extractBinary(tmpDir, archivePath, c.App)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	target, err := currentExecutablePath()
	if err != nil {
		return err
	}
	backup := target + ".old"

	if runtime.GOOS == "windows" {
		_, _ = fmt.Fprintln(out, "Applying update...")
		return applyWindows(target, backup, newBinPath, out, errOut)
	}

	_, _ = fmt.Fprintln(out, "Applying update...")
	if err := applyInPlace(target, backup, newBinPath); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "✓ Updated to %s (restart `kairo`)\n", displayVersion(res.Latest))
	return nil
}

func displayVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "dev" {
		return "dev"
	}
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}

func needsUpdate(current, latest string) (bool, error) {
	cv, err := semver.NewVersion(current)
	if err != nil {
		cv = semver.MustParse("0.0.0")
	}
	lv, err := semver.NewVersion(latest)
	if err != nil {
		return false, fmt.Errorf("invalid latest version %q", latest)
	}
	return lv.GreaterThan(cv), nil
}

func sha256FileHex(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func applyInPlace(target, backup, newBinPath string) error {
	f, err := os.Open(newBinPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	opts := selfupdate.Options{
		TargetPath:  target,
		TargetMode:  0o755,
		OldSavePath: backup,
	}
	if err := opts.CheckPermissions(); err != nil {
		return err
	}
	if err := selfupdate.Apply(f, opts); err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("update failed and rollback failed: %v (rollback: %v)", err, rerr)
		}
		return err
	}
	return nil
}

func applyWindows(target, backup, newBinPath string, stdout, stderr io.Writer) error {
	helper, err := stageWindowsHelper()
	if err != nil {
		return err
	}

	args := []string{
		"__kairo_selfupdate_apply",
		"--target", target,
		"--backup", backup,
		"--source", newBinPath,
	}

	cmd := exec.Command(helper, args...)
	// On Windows, we need to detach or at least not wait for it.
	// We'll use os.StartProcess style via cmd.Start() but then we MUST exit.
	if err := cmd.Start(); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(stdout, "Update staged. Closing to finish update...")
	os.Exit(0)
	return nil
}

func stageWindowsHelper() (string, error) {
	exe, err := currentExecutablePath()
	if err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp("", "kairo-selfupdate-*")
	if err != nil {
		return "", err
	}

	helper := filepath.Join(dir, "kairo-selfupdate.exe")
	if err := copyFile(exe, helper, 0o755); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	return helper, nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func currentExecutablePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	ex, err = filepath.EvalSymlinks(ex)
	if err != nil {
		return "", err
	}
	return ex, nil
}

func expectedArchiveName(app string) string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}
	base := fmt.Sprintf("%s_%s_%s", app, runtime.GOOS, arch)
	if runtime.GOOS == "windows" {
		return base + ".zip"
	}
	return base + ".tar.gz"
}

func findAsset(assets []Asset, name string) (Asset, bool) {
	for _, a := range assets {
		if a.Name == name {
			return a, true
		}
	}
	return Asset{}, false
}
