package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/programmersd21/kairo/internal/api"
	"github.com/programmersd21/kairo/internal/app"
	"github.com/programmersd21/kairo/internal/buildinfo"
	"github.com/programmersd21/kairo/internal/completion"
	"github.com/programmersd21/kairo/internal/config"
	"github.com/programmersd21/kairo/internal/core"
	"github.com/programmersd21/kairo/internal/core/codec"
	"github.com/programmersd21/kairo/internal/hooks"
	"github.com/programmersd21/kairo/internal/mcp"
	"github.com/programmersd21/kairo/internal/service"
	"github.com/programmersd21/kairo/internal/storage"
	ksync "github.com/programmersd21/kairo/internal/sync"
	"github.com/programmersd21/kairo/internal/updater"
)

func main() {
	if handled, err := updater.MaybeRunWindowsApply(os.Stdout, os.Stderr); handled {
		if err != nil {
			fmt.Fprintln(os.Stderr, "kairo update:", err)
			os.Exit(2)
		}
		return
	}

	// Immediate subcommands (no config/DB needed)
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "completion":
			if len(os.Args) < 3 {
				fmt.Println("Usage: kairo completion [bash|zsh|fish|powershell] [install]")
				fmt.Println("       kairo completion --complete [args...]")
				os.Exit(1)
			}
			if os.Args[2] != "--complete" {
				shell := os.Args[2]
				if len(os.Args) > 3 && os.Args[3] == "install" {
					if err := completion.Install(shell); err != nil {
						fmt.Fprintln(os.Stderr, "Error:", err)
						os.Exit(1)
					}
					return
				}
				script, err := completion.Script(shell)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				fmt.Print(script)
				return
			}
		case "version":
			runVersion()
			return
		case "help":
			runHelp(os.Args[2:])
			return
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kairo: config:", err)
		os.Exit(2)
	}

	repo, err := storage.Open(ctx, cfg.Storage.Path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kairo: storage:", err)
		os.Exit(2)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "kairo: failed to close storage:", err)
		}
	}()

	// Initialize unified service layer
	hks := hooks.New()
	svc := service.New(repo, hks)

	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "completion":
			if os.Args[2] == "--complete" {
				results := completion.Complete(ctx, svc, os.Args[3:])
				for _, r := range results {
					fmt.Println(r)
				}
				return
			}
		case "api":
			if err := runAPI(ctx, svc, os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, "kairo api:", err)
				os.Exit(2)
			}
			return
		case "export":
			if err := runExport(ctx, repo, os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, "kairo export:", err)
				os.Exit(2)
			}
			return
		case "import":
			if err := runImport(ctx, repo, os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, "kairo import:", err)
				os.Exit(2)
			}
			return
		case "sync":
			if !cfg.Sync.Enabled || strings.TrimSpace(cfg.Sync.RepoPath) == "" {
				fmt.Fprintln(os.Stderr, "kairo sync: enable sync.repo_path in config")
				os.Exit(2)
			}
			eng := ksync.New(repo, cfg.Sync.RepoPath, cfg.Sync.Remote, cfg.Sync.Branch, ksync.Strategy(cfg.Sync.Strategy), cfg.Sync.AutoPush)
			if err := eng.SyncNow(ctx); err != nil {
				fmt.Fprintln(os.Stderr, "kairo sync:", err)
				os.Exit(2)
			}
			return
		case "update":
			if err := runUpdate(ctx); err != nil {
				fmt.Fprintln(os.Stderr, "kairo update:", err)
				os.Exit(2)
			}
			return
		case "mcp":
			if err := runMCP(ctx, svc, os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, "kairo mcp:", err)
				os.Exit(2)
			}
			return
		}
	}

	// Emit app start event (plugins can listen to this)
	hks.AppStarted()

	// Perform database cleanup on startup
	if err := svc.Prune(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "kairo: database cleanup:", err)
	}

	defer hks.AppStopped()

	m, err := app.New(ctx, cfg, svc)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kairo:", err)
		os.Exit(2)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithContext(ctx))
	if _, err := p.Run(); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "kairo:", err)
		os.Exit(1)
	}
}

func runAPI(ctx context.Context, svc service.TaskService, args []string) error {
	if len(args) == 0 {
		return errors.New("missing action (create, list, update, delete, get, list-tags, cleanup)")
	}

	taskAPI := api.New(svc)
	action := args[0]

	// Handle "delete all" special case
	if action == "delete" && len(args) > 1 && args[1] == "all" {
		action = "delete_all"
		args = args[1:]
	}

	// Handle "configure-ai" set/reset
	if action == "configure-ai" {
		payload := make(map[string]interface{})
		if len(args) > 1 {
			if args[1] == "set" && len(args) > 2 {
				payload["key"] = args[2]
			} else if args[1] == "reset" {
				payload["reset"] = true
			}
		}
		b, _ := json.Marshal(payload)
		resp := taskAPI.Execute(ctx, api.Request{Action: action, Payload: b})
		out, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(out))
		return nil
	}

	var req api.Request
	if action == "--json" {
		if len(args) < 2 {
			return errors.New("--json requires a JSON string")
		}
		if err := json.Unmarshal([]byte(args[1]), &req); err != nil {
			return fmt.Errorf("invalid json: %w", err)
		}
	} else {
		req.Action = action
		payload := make(map[string]interface{})
		for i := 1; i < len(args); i++ {
			if strings.HasPrefix(args[i], "--") {
				key := strings.TrimPrefix(args[i], "--")
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
					payload[key] = args[i+1]
					i++
				} else {
					payload[key] = true
				}
			}
		}
		b, _ := json.Marshal(payload)
		req.Payload = b
	}

	resp := taskAPI.Execute(ctx, req)
	out, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(out))
	return nil
}

func runExport(ctx context.Context, repo *storage.Repository, args []string) error {
	format := "json"
	out := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = strings.ToLower(args[i+1])
				i++
			}
		case "--out":
			if i+1 < len(args) {
				out = args[i+1]
				i++
			}
		}
	}
	tasks, err := repo.AllTasks(ctx)
	if err != nil {
		return err
	}

	var b []byte
	switch format {
	case "csv":
		b, err = codec.MarshalCSV(tasks)
	case "txt", "text":
		b = codec.MarshalText(tasks)
	case "md", "markdown":
		b = codec.MarshalMarkdown(tasks)
	case "json":
		b, err = codec.MarshalJSON(tasks)
	default:
		return fmt.Errorf("unknown format %q (supported: json, md, csv, txt)", format)
	}
	if err != nil {
		return err
	}

	if out == "" {
		_, err = os.Stdout.Write(b)
		if err == nil && len(b) > 0 && b[len(b)-1] != '\n' {
			_, _ = os.Stdout.Write([]byte("\n"))
		}
		return err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(out, b, 0o644)
}

func runImport(ctx context.Context, repo *storage.Repository, args []string) error {
	format := "json"
	in := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = strings.ToLower(args[i+1])
				i++
			}
		case "--in":
			if i+1 < len(args) {
				in = args[i+1]
				i++
			}
		}
	}
	if in == "" {
		return errors.New("--in required")
	}
	b, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	var tasks []core.Task
	switch format {
	case "csv":
		tasks, err = codec.UnmarshalCSV(b)
	case "txt", "text":
		tasks, err = codec.UnmarshalText(b)
	case "md", "markdown":
		tasks, err = codec.UnmarshalMarkdown(b)
	case "json":
		tasks, err = codec.UnmarshalJSON(b)
	default:
		return fmt.Errorf("unknown format %q (supported: json, md, csv, txt)", format)
	}
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if err := repo.UpsertTask(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func runVersion() {
	fmt.Printf("kairo %s\n", buildinfo.VersionWithCommit())
}

func runUpdate(ctx context.Context) error {
	cfg := updater.DefaultConfig()
	return cfg.Update(ctx, updater.UpdateOptions{
		CurrentVersion: buildinfo.EffectiveVersion(),
		Stdout:         os.Stdout,
		Stderr:         os.Stderr,
	})
}

func runHelp(args []string) {
	if len(args) == 0 {
		fmt.Println("Kairo — Minimal, powerful task management.")
		fmt.Println("\nUsage:")
		fmt.Println("  kairo [command]")
		fmt.Println("\nAvailable Commands:")
		fmt.Println("  api         Headless API for external automation")
		fmt.Println("  completion  Generate shell completion scripts")
		fmt.Println("  export      Export tasks to JSON or Markdown")
		fmt.Println("  import      Import tasks from JSON or Markdown")
		fmt.Println("  mcp         Run the built-in MCP server")
		fmt.Println("  sync        Sync tasks with Git repository")
		fmt.Println("  update      Update Kairo to the latest version")
		fmt.Println("  version     Show the current version")
		fmt.Println("  help        Help about any command")
		fmt.Println("\nUse \"kairo help [command]\" for more information about a command.")
		return
	}

	switch args[0] {
	case "api":
		fmt.Println("Headless API for external automation.")
		fmt.Println("\nUsage:")
		fmt.Println("  kairo api [action] [flags]")
		fmt.Println("\nActions:")
		fmt.Println("  create, list, update, delete, delete all, get, list-tags, cleanup")
	case "completion":
		fmt.Println("Generate shell completion scripts.")
		fmt.Println("\nUsage:")
		fmt.Println("  kairo completion [bash|zsh|fish|powershell] [install]")
		fmt.Println("  kairo completion --complete [args...]")
		fmt.Println("\nExample:")
		fmt.Println("  kairo completion zsh install")
	case "export":
		fmt.Println("Export tasks to JSON, Markdown, CSV or Text.")
		fmt.Println("\nUsage:")
		fmt.Println("  kairo export --format [json|md|csv|txt] --out [file]")
	case "import":
		fmt.Println("Import tasks from JSON, Markdown, CSV or Text.")
		fmt.Println("\nUsage:")
		fmt.Println("  kairo import --format [json|md|csv|txt] --in [file]")
	case "sync":
		fmt.Println("Sync tasks with Git repository.")
		fmt.Println("\nUsage:")
		fmt.Println("  kairo sync")
	case "mcp":
		fmt.Println("Run the built-in Model Context Protocol (MCP) server.")
		fmt.Println("\nUsage:")
		fmt.Println("  kairo mcp [port]")
		fmt.Println("\nExample:")
		fmt.Println("  kairo mcp        (Runs in Stdio mode for local AI agents)")
		fmt.Println("  kairo mcp 8080   (Runs in SSE/HTTP mode on port 8080)")
		fmt.Println("\nDescription:")
		fmt.Println("  Exposes your Kairo tasks and tools to AI agents using the MCP standard.")
		fmt.Println("  Stdio mode is used for local integration (e.g. Claude Desktop).")
		fmt.Println("  SSE mode allows remote connections via HTTP.")
	default:
		fmt.Printf("Unknown help topic %q\n", args[0])
	}
}
func runMCP(ctx context.Context, svc service.TaskService, args []string) error {
	s := mcp.NewServer(svc)

	cfg, _ := config.Load()
	port := os.Getenv("KAIRO_MCP_PORT")
	if port == "" {
		port = cfg.App.MCPPort
	}
	if len(args) > 0 {
		port = args[0]
	}

	if port != "" {
		addr := ":" + port
		baseURL := "http://localhost" + addr
		sseServer := mcpserver.NewSSEServer(s, mcpserver.WithBaseURL(baseURL))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			sseServer.ServeHTTP(w, r)
		})

		fmt.Printf("Starting Kairo MCP SSE server on %s (CORS enabled)\n", addr)
		fmt.Printf("SSE endpoint: %s/sse\n", baseURL)
		return http.ListenAndServe(addr, handler)
	}

	server := mcpserver.NewStdioServer(s)
	return server.Listen(ctx, os.Stdin, os.Stdout)
}
