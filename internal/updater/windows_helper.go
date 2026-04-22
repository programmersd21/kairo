package updater

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/minio/selfupdate"
)

func MaybeRunWindowsApply(stdout, stderr io.Writer) (handled bool, err error) {
	if runtime.GOOS != "windows" {
		return false, nil
	}
	if len(os.Args) < 2 || os.Args[1] != "__kairo_selfupdate_apply" {
		return false, nil
	}

	fs := flag.NewFlagSet("__kairo_selfupdate_apply", flag.ContinueOnError)
	fs.SetOutput(stderr)
	target := fs.String("target", "", "target exe")
	backup := fs.String("backup", "", "backup exe")
	source := fs.String("source", "", "new exe")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return true, err
	}
	if *target == "" || *source == "" {
		return true, errors.New("--target and --source are required")
	}
	if *backup == "" {
		*backup = *target + ".old"
	}

	// Give parent process time to exit
	time.Sleep(1 * time.Second)

	if err := applyWithRetry(*target, *backup, *source, 30*time.Second); err != nil {
		// Since we're in a detached-like process, we might want to log to a file
		// but for now, we'll try to use stderr if it's still connected.
		return true, fmt.Errorf("failed to apply update: %w", err)
	}

	_ = os.Remove(*source)
	// We don't necessarily need to delete ourselves if we're in a temp dir,
	// but it's cleaner.
	scheduleSelfDelete()
	return true, nil
}

func applyWithRetry(target, backup, source string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		err := func() error {
			f, err := os.Open(source)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()

			err = selfupdate.Apply(f, selfupdate.Options{
				TargetPath:  target,
				TargetMode:  0o755,
				OldSavePath: backup,
			})
			if rerr := selfupdate.RollbackError(err); rerr != nil {
				return fmt.Errorf("update failed and rollback failed: %v (rollback: %v)", err, rerr)
			}
			return err
		}()
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return err
		}
		time.Sleep(750 * time.Millisecond)
	}
}

func scheduleSelfDelete() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	exe, _ = filepath.EvalSymlinks(exe)
	if exe == "" {
		return
	}

	_ = exec.Command("cmd", "/C", fmt.Sprintf("timeout /t 2 /nobreak >NUL & del /F /Q %q", exe)).Start()
}
