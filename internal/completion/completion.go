package completion

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/programmersd21/kairo/internal/service"
	"github.com/programmersd21/kairo/internal/util"
)

// Shells supported for completion
const (
	Bash       = "bash"
	Zsh        = "zsh"
	Fish       = "fish"
	PowerShell = "powershell"
)

// Scripts returns the completion script for the given shell
func Script(shell string) (string, error) {
	switch shell {
	case Bash:
		return bashScript, nil
	case Zsh:
		return zshScript, nil
	case Fish:
		return fishScript, nil
	case PowerShell:
		return powerShellScript, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
	}
}

// Install adds the completion script to the shell's profile
func Install(shell string) error {
	appDir, err := util.AppDataDir("kairo")
	if err != nil {
		return err
	}
	_ = os.MkdirAll(appDir, 0755)

	script, err := Script(shell)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var profilePath string
	var sourceCmd string
	scriptPath := filepath.Join(appDir, "completion."+shell)

	switch shell {
	case Bash:
		profilePath = filepath.Join(home, ".bashrc")
		sourceCmd = fmt.Sprintf("\n# kairo completion\nif [ -f %q ]; then . %q; fi\n", scriptPath, scriptPath)
	case Zsh:
		profilePath = filepath.Join(home, ".zshrc")
		sourceCmd = fmt.Sprintf("\n# kairo completion\nif [ -f %q ]; then . %q; fi\n", scriptPath, scriptPath)
	case Fish:
		configDir := filepath.Join(home, ".config", "fish", "completions")
		_ = os.MkdirAll(configDir, 0755)
		profilePath = filepath.Join(configDir, "kairo.fish")
		return os.WriteFile(profilePath, []byte(script), 0644)
	case PowerShell:
		path, err := getPowerShellProfile()
		if err != nil {
			return err
		}
		profilePath = path
		sourceCmd = fmt.Sprintf("\n# kairo completion\nif (Test-Path %q) { . %q }\n", scriptPath, scriptPath)
		scriptPath = filepath.Join(appDir, "completion.ps1")
	default:
		return fmt.Errorf("unsupported shell for automatic installation: %s", shell)
	}

	// Write the static script file
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("could not write completion script: %w", err)
	}

	// If it's fish, we're done (we wrote to the completions dir directly)
	if shell == Fish {
		fmt.Printf("✓ Successfully installed %s completions to %s\n", shell, profilePath)
		return nil
	}

	// For others, append source command to profile
	content, _ := os.ReadFile(profilePath)
	if strings.Contains(string(content), "kairo completion") {
		fmt.Printf("✓ Completion script updated in %s (source command already exists in %s)\n", scriptPath, profilePath)
		return nil
	}

	f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if _, err := f.WriteString(sourceCmd); err != nil {
		_ = f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	fmt.Printf("✓ Successfully installed %s completions to %s\n", shell, profilePath)
	fmt.Println("  Please restart your shell or run: source " + profilePath)
	return nil
}

func getPowerShellProfile() (string, error) {
	// Try pwsh first, then powershell
	for _, exe := range []string{"pwsh", "powershell"} {
		out, err := exec.Command(exe, "-NoProfile", "-Command", "$PROFILE").Output()
		if err == nil {
			path := strings.TrimSpace(string(out))
			if path != "" {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("could not detect PowerShell profile path (is PowerShell installed?)")
}

// Complete returns suggestions based on the current command line
func Complete(ctx context.Context, svc service.TaskService, args []string) []string {
	// args[0] is usually the program name, we want what's after
	if len(args) <= 1 {
		return []string{"api", "export", "import", "sync", "version", "update", "completion"}
	}

	subcommand := args[1]

	// If we are still typing the first subcommand
	if len(args) == 2 {
		choices := []string{"api", "export", "import", "sync", "version", "update", "completion"}
		return filterPrefix(choices, subcommand)
	}

	switch subcommand {
	case "api":
		return completeAPI(ctx, svc, args[2:])
	case "export":
		return completeExport(args[2:])
	case "import":
		return completeImport(args[2:])
	case "completion":
		return []string{Bash, Zsh, Fish, PowerShell}
	}

	return nil
}

func completeAPI(ctx context.Context, svc service.TaskService, args []string) []string {
	actions := []string{"create", "list", "update", "delete", "get", "list-tags"}
	if len(args) <= 1 {
		prefix := ""
		if len(args) == 1 {
			prefix = args[0]
		}
		return filterPrefix(actions, prefix)
	}

	action := args[0]
	// Suggest task IDs for actions that need them
	if action == "get" || action == "delete" || action == "update" {
		tasks, _ := svc.ListAll(ctx)
		var ids []string
		for _, t := range tasks {
			ids = append(ids, t.ID)
		}
		prefix := ""
		if len(args) > 1 {
			prefix = args[len(args)-1]
		}
		return filterPrefix(ids, prefix)
	}

	return nil
}

func completeExport(args []string) []string {
	flags := []string{"--format", "--out"}
	if len(args) == 0 {
		return flags
	}

	last := args[len(args)-1]
	prev := ""
	if len(args) > 1 {
		prev = args[len(args)-2]
	}

	if prev == "--format" {
		return filterPrefix([]string{"json", "md", "markdown"}, last)
	}

	if strings.HasPrefix(last, "-") {
		return filterPrefix(flags, last)
	}

	return flags
}

func completeImport(args []string) []string {
	flags := []string{"--format", "--in"}
	if len(args) == 0 {
		return flags
	}
	last := args[len(args)-1]
	if strings.HasPrefix(last, "-") {
		return filterPrefix(flags, last)
	}
	return nil
}

func filterPrefix(choices []string, prefix string) []string {
	var filtered []string
	for _, c := range choices {
		if strings.HasPrefix(c, prefix) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

const bashScript = `
_kairo_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$(kairo completion --complete "${COMP_WORDS[@]:0:$COMP_CWORD}" "$cur")
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}
complete -F _kairo_completions kairo
`

const zshScript = `
#compdef kairo
_kairo() {
    local -a opts
    opts=(${(f)"$(kairo completion --complete ${words[@]:0:${#words}-1} "${words[CURRENT]}")"})
    _describe 'values' opts
}
compdef _kairo kairo
`

const fishScript = `
complete -c kairo -f -a "(kairo completion --complete (commandline -opc) (commandline -ct))"
`

const powerShellScript = `
Register-ArgumentCompleter -CommandName kairo -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $args = $commandAst.ToString().Split(' ')
    kairo completion --complete $args | Where-Object { $_ -like "$wordToComplete*" }
}
`
