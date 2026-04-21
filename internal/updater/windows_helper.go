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

	_, _ = fmt.Fprintln(stdout, "Finishing update...")
	if err := applyWithRetry(*target, *backup, *source, 45*time.Second); err != nil {
		return true, err
	}

	_ = os.Remove(*source)
	scheduleSelfDelete()
	_, _ = fmt.Fprintln(stdout, "✓ Update applied. Re-run `kairo`.")
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
