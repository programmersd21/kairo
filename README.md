<div align="center">

# 📝 Kairo

### 🌿 Minimal, powerful task management for the modern terminal.

![Demo](screenshots/demo.gif)

[![Release](https://img.shields.io/github/v/release/programmersd21/kairo?sort=semver&style=for-the-badge&logo=github&color=7c3aed)](https://github.com/programmersd21/kairo/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/programmersd21/kairo/ci.yml?branch=main&style=for-the-badge&logo=githubactions&logoColor=white&color=2563eb)](https://github.com/programmersd21/kairo/actions)
[![Go Report Card](https://img.shields.io/badge/go%20report-A%2B-brightgreen?style=for-the-badge&logo=go&logoColor=white&color=10b981)](https://goreportcard.com/report/github.com/programmersd21/kairo)
[![License: MIT](https://img.shields.io/badge/License-MIT-f59e0b?style=for-the-badge&logo=open-source-initiative&logoColor=white)](https://opensource.org/licenses/MIT)

**⌛ Time, executed well.**

</div>

---

### ✨ A Premium Terminal Task Manager Designed for Focused Execution

🏃🏻 **Kairo** is a *lightning-fast*, **keyboard-first** task management application  
built for developers and power users.

It combines the simplicity of a **command-line tool**  
with the sophistication of a *modern, premium design system*.

🎯 **BubbleTea Motion System** — Liquid glass interactions with elastic physics  
🎨 **Premium UI Design** — Modern Bento-style layout with soft, rounded aesthetics  
⌨️ **Keyboard-First** — Complete control without ever touching a mouse  
🖥️ **Seamless Rendering** — Pixel-perfect background fills the entire viewport, no terminal bleed-through  
🔐 **Offline-First** — Your data lives locally in SQLite, always under your control  
🔗 **Git-Backed Sync** — Optional distributed sync leveraging Git's architecture  
🧩 **Extensible** — Unified Lua plugin system and CLI automation API  
📱 **Responsive Layout** — Gracefully adapts to any terminal size  
🤖 **Automation-Friendly** — Headless API for external scripts and CI/CD  
🌊 **Boba Liquid Feel** — UI elements behave with soft inertia and fluid clustering  

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) (TUI framework), [Lip Gloss](https://github.com/charmbracelet/lipgloss) (terminal styling), and SQLite (local storage).

---

## ✨ Core Features

| Feature | Description |
|---------|-------------|
| **Task Service** | Single source of truth for TUI, Lua, and CLI automation |
| **Lua Plugins** | Native first-class scripting with event hooks (GopherLua) |
| **Automation API** | Stable CLI interface for external control and JSON integration |
| **Event Hooks** | React to task creation, updates, and app lifecycle events |
| **Smart Filtering** | Multiple views: Inbox, Today, Upcoming, Completed, by Tag, by Priority |
| **Fuzzy Search** | Lightning-fast command palette with ranked results |
| **Cinematic Animations** | Smooth vertical shutter, cascading row reveals, and glitch/vaporize deletions |
| **Responsive Auto-Resize**| Strict grid enforcement with dynamic title truncation preventing layout drifts |
| **Offline Storage** | SQLite with WAL for reliability and concurrent access |
| **Git Sync** | Optional repository-backed sync with per-task JSON files |
| **Import/Export** | JSON, Markdown, CSV, and Text support for data portability |
| **AI Assistant** | Integrated Gemini (3.1/2.5/2.0) with total app control, Google Search, & live UI refreshes |
| **MCP Server**   | Built-in Model Context Protocol server exposing entire task schema, themes, and plugins |
| **Custom Themes**| Curate and share custom themes via Lua plugins or API |

---

## 🤩 Star History

<a href="https://www.star-history.com/?repos=programmersd21%2Fkairo&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=programmersd21/kairo&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=programmersd21/kairo&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=programmersd21/kairo&type=date&legend=top-left" />
 </picture>
</a>

---

## 📦 Installation

### macOS (Homebrew)

```bash
brew tap programmersd21/kairo_tap
brew install --cask kairo
```

### Linux / macOS (curl)

```bash
curl -fsSL https://raw.githubusercontent.com/programmersd21/kairo/main/scripts/install.sh | bash
```

Installs to `$HOME/.local/bin/kairo` (fallback: `/usr/local/bin/kairo`) and attempts to persist the PATH update in your shell profile when needed.

### Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/programmersd21/kairo/main/scripts/install.ps1 | iex
```

Installs to `%USERPROFILE%\AppData\Local\Programs\kairo\kairo.exe` and adds the install directory to your user PATH.

### Any other OS

```bash
go install github.com/programmersd21/kairo/cmd/kairo@latest
```

**OR** download a prebuilt binary from the [Releases page](https://github.com/programmersd21/kairo/releases).

### Updates

```bash
kairo update
```

Downloads the latest GitHub Release for your OS/arch, verifies it against `checksums.txt`, and safely replaces the installed binary.
On Windows, Kairo will automatically close to apply the update; simply re-run `kairo` once the terminal returns.

**Startup Notifications:**
Kairo automatically checks for updates on startup. If a newer version is available, a notification will appear in the footer (e.g., `Update: v1.2.2 → v1.2.3`) directing you to run the update command.

---

## 🤖 Automation & CLI API

Kairo provides a stable CLI API for external automation. Every operation available in the TUI can be performed via the `api` subcommand.

### Usage

```bash
# List tasks with a specific tag
kairo api list --tag work

# Create a new task
kairo api create --title "Finish report" --priority 1

# Update a task
kairo api update --id <task-id> --status done

# Delete all tasks (soft-delete)
kairo api delete all

# Advanced JSON interface
kairo api --json '{"action": "create", "payload": {"title": "API Task", "tags": ["bot"]}}'

# AI Configuration
kairo api configure-ai set "YOUR_GEMINI_API_KEY"
kairo api configure-ai reset

# Set TUI Theme
kairo api set_theme --theme catppuccin

# Plugin Management (list, get, write, delete)
kairo api plugin_list
kairo api plugin_get --name auto-cleanup.lua
kairo api plugin_delete --name sample.lua
```

### Other CLI Commands

```bash
# Check installed version
kairo version

# Update to the latest version
kairo update

# Export tasks
kairo export --format csv --out tasks.csv
kairo export --format txt --out tasks.txt
kairo export --format markdown --out tasks.md

# Import tasks
kairo import --format json --in tasks.json

# Shell completion (bash, zsh, fish, powershell)
# Automatic install:
kairo completion zsh install

# Manual install (add to your shell profile):
# source <(kairo completion zsh)
kairo completion zsh

# Get help for any command
kairo help
kairo help api
kairo help export

# Sync with Git (if configured)
kairo sync

# Start MCP Server (stdio)
kairo mcp
```

---

## 🔌 Plugins (Lua)

Extend Kairo with custom logic, event hooks, commands, and views using Lua.

### Plugin Structure

```lua
-- plugins/my-plugin.lua
local plugin = {
    id = "my-plugin",
    name = "My Plugin",
    description = "Reacts to tasks",
    version = "1.0.0",
}

-- Hook into events
kairo.on("task_create", function(event)
    kairo.notify("New task: " .. event.task.title)
end)

-- Register custom commands
plugin.commands = {
    { id = "hello", title = "Say Hello", run = function() kairo.notify("Hello!") end }
}


-- Register custom themes
plugin.themes = {
    {
        name = "midnight_neon",
        is_light = false,
        bg = "#000000",
        fg = "#ffffff",
        muted = "#444444",
        border = "#222222",
        accent = "#00ff00",
        good = "#00ff00",
        warn = "#ffff00",
        bad = "#ff0000",
        overlay = "#111111",
    }
}

return plugin
```

### Supported Events
- `task_create`
- `task_update`
- `task_delete`
- `app_start`
- `app_stop`

### Lua API Reference

| Method | Description |
|--------|-------------|
| `kairo.create_task(table)` | Create a new task |
| `kairo.update_task(id, table)` | Update an existing task |
| `kairo.delete_task(id)` | Delete a task |
| `kairo.list_tasks(filter)` | List tasks with optional filter |
| `kairo.on(event, function)` | Register an event listener |
| `kairo.notify(msg, is_error)` | Send a notification to the UI |

---

## 🎨 Design System

Kairo features a **minimalist design system** optimized for clarity and focus.

### Design Philosophy

- **Breathable Layout** — Reduced padding and thin borders for a clean, modern look
- **Seamless Backdrop** — Custom rendering engine ensures the theme background covers the entire terminal window
- **Instant Feedback** — Smooth strikethrough animations when completing tasks
- **Keyboard-First** — All interactions optimized for speed
- **High Compatibility** — Uses standard Unicode symbols for consistent rendering across all terminals

---

## ⌨️ Keyboard Navigation

### Essential Commands

| Shortcut | Action |
|----------|--------|
| `ctrl+p` | 🔍 Open Command Palette |
| `z` | ⚡ **Strike (toggle completion with animation)** |
| `tab` / `shift+tab` | → / ← Switch views |
| `n` | ➕ Create new task |
| `e` | ✏️ Edit selected task |
| `enter` | 👁️ View task details |
| `d` | 🗑️ Delete task |
| `t` | 🎨 Cycle themes |
| `ctrl+s` | ⚙️ Open Settings Menu |
| `i` | 📢 Open GitHub issues |
| `c` | 📝 Show changelog |
| `?` | ❓ Show help menu |
| `q` | ❌ Quit |

### AI Assistant Shortcuts

| Shortcut | Action |
|----------|--------|
| `ctrl+a` | 🤖 Toggle AI Assistant Panel |
| `ctrl+l` | 🧹 Clear AI Chat History |
| `enter`  | ↵ Submit Prompt |
| `esc`    | ❌ Close AI Panel |

### Plugin Menu Shortcuts

| Shortcut | Action |
|----------|--------|
| `enter` | 👁️ View plugin details |
| `u` | 🗑️ Uninstall plugin |
| `o` | 📂 Open plugins folder |
| `r` | 🔄 Reload plugins |
| `p` / `esc` | ❌ Close menu |

### View Shortcuts

| Shortcut | View |
|----------|------|
| `1` - `9` | **Switch Views** — Instant access to all tabs (Inbox, Today, Plugins, etc.) |
| `f` | **Tag Filter** — Quickly jump to Tag View and filter by one or multiple tags (e.g., `work dev kairo`) |
| `tab` / `shift+tab` | **Cycle Views** — Move through all available tabs |

### Pro Tips
- Press `f` to open the **tag filter input modal** for direct tag entry
- Type tag name and press `enter` to apply filter, or `esc` to cancel
- Type `#tag` in the command palette to jump to a specific tag
- Type `pri:0` to filter tasks by priority level
- Use `ctrl+s` to save while editing
- Press `esc` to cancel and return to the list

---

## ⌨️ Vim Mode

For users who live in the terminal, Kairo offers a built-in **Vim Mode** for seamless navigation without leaving the home row.

### Enabling Vim Mode
You can toggle Vim Mode in two ways:
1. **Settings Menu**: Press `ctrl+s` and toggle "Vim Mode" to `true`.
2. **Configuration File**: Set `vim_mode = true` in your `config.toml`.

### Vim Shortcuts
When enabled, the following classic Vim keys are activated for list navigation:
- `j`: Move selection down
- `k`: Move selection up
- `G`: Jump to the bottom of the list
- `gg`: Jump to the top of the list

*Note: Standard arrow keys, `pgup`/`pgdown`, and `home`/`end` remain functional regardless of this setting.*

---

## ⚙️ Configuration

### Config Location

| OS | Path |
|----|------|
| **Windows** | `%APPDATA%\kairo\config.toml` |
| **macOS** | `~/Library/Application Support/kairo/config.toml` |
| **Linux** | `~/.config/kairo/config.toml` |

### Quick Setup

```bash
cp configs/kairo.example.toml ~/.config/kairo/config.toml
```

Then edit the file or use the built-in Settings menu (`ctrl+s` in Kairo) to customize:
- **Theme selection** — Choose from 32 built-in themes:
    - **Premium Dark:** `catppuccin` (Default), `midnight`, `aurora`, `cyberpunk`, `dracula`, `nord`, `obsidian_bloom`, `neon_reef`, `carbon_sunset`, `vanta_aurora`, `plasma_grape`, `midnight_jade`, `synthwave_minimal`, `graphite_matcha`
    - **Premium Light:** `vanilla`, `solarized`, `rose`, `matcha`, `cloud`, `sepia`, `cloud_dancer`, `sakura_sand`, `olive_mist`, `terracotta_air`, `vanilla_sky`, `peach_fuzz_neo`, `coastal_drift`, `matcha_latte`
    - **Hybrid/Specialized:** `digital_lavender`, `neo_mint_system`, `sunset_gradient_pro`, `forest_sanctuary`
- **AI Model Selection** — Switch between `gemini-3.1-flash-lite-preview`, `gemini-2.5-flash-lite`, and `gemini-2.0-flash-lite` live using ←/→ arrows
- **Keybindings** — Rebind any keyboard shortcut
- **View ordering** — Customize your task view tabs
- **Sync settings** — Configure Git repository sync
- **Plugins** — Toggle and manage your Lua plugins
- **Reset to Defaults** — Press `r` inside the Settings menu to restore all factory settings

---

## 🔄 Git Sync

Enable optional distributed sync by setting `sync.repo_path` in your config.

Kairo uses a unique no-backend approach:
- Each task is stored as an individual JSON file
- Changes are committed locally with automatic messages
- Perform sync manually or on-demand
- Git's branching and merging handle conflicts transparently

```bash
# Manual sync
kairo sync
```

---

## 📅 Natural Language Deadlines

Kairo's smart parser understands natural language, making it effortless to set deadlines without worrying about specific date formats.

When creating or editing a task, you can input deadlines like:
- **Relative days:** `today`, `tomorrow`, `day after tomorrow`
- **Specific days:** `monday`, `next friday`, `this sunday`
- **Time-based:** `in 2 hours`, `at 5pm`, `tomorrow at 10am`
- **Dates:** `august 24`, `24th of april`

Powered by the [when](https://github.com/olebedev/when) library, Kairo ensures your deadlines are always parsed intuitively.

---

## 🏗 Architecture

Kairo is built with a modular architecture designed for performance, extensibility, and data sovereignty.

### Core Components

| Component | Role |
|-----------|------|
| **Task Service** | Single source of truth for TUI, Lua, and CLI automation |
| **UI Layer** ([Bubble Tea](https://github.com/charmbracelet/bubbletea)) | Elm-inspired TUI framework with state-machine pattern for mode management |
| **Storage** (SQLite) | Pure Go database with WAL for reliability and concurrent access |
| **Sync Engine** (Git) | Distributed "no-backend" sync with per-task JSON files |
| **Search** (Fuzzy Index) | In-memory ranked matching with sub-millisecond results |
| **Plugins** ([Gopher-Lua](https://github.com/yuin/gopher-lua)) | Lightweight Lua VM for user extensions |

### Data Flow

```
User Input/API/Lua → Task Service → Hooks System
    ↓
Immediate DB Persistence → Optional Git Sync
    ↓
UI Re-render → Instant User Feedback
```

---

## 🌴 Project Structure

```
kairo/
├── cmd/
│   └── kairo/
│       └── main.go            # Entry point for TUI & CLI
├── configs/
│   └── kairo.example.toml     # Template configuration
├── internal/
│   ├── ai/                    # Gemini API & Tool-calling engine
│   ├── api/                   # Headless JSON API & Plugin control
│   ├── app/                   # Root TUI state & message bus
│   ├── core/                  # Task models & NLP logic
│   │   └── codec/             # CSV, JSON, Markdown, Text support
│   ├── mcp/                   # Model Context Protocol server
│   ├── plugins/               # Lua plugin host (host.go)
│   ├── storage/               # SQLite & Migration engine
│   ├── sync/                  # Optional Git-backed sync logic
│   ├── ui/                    # Componentized TUI (Bubble Tea)
│   │   ├── ai_panel/          # Integrated AI assistant
│   │   ├── import_export_menu/# Format-aware I/O interface
│   │   ├── settings/          # Live configuration & model switching
│   │   └── ...
│   └── util/                  # Cross-platform path helpers
├── plugins/                   # User-extensible Lua scripts
├── screenshots/               # Demo assets
└── scripts/                   # Platform-specific installers
```

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for our code of conduct.

### Areas for Contribution
- ✨ New themes and design improvements
- 🐛 Bug fixes and performance enhancements
- 📚 Documentation and tutorials
- 🧩 Plugins and extensions
- 🌍 Translations and localization

---

## 💙 Community Legend(s)

- **@Tornado300** — Contributed significantly by reporting issues that led to multiple critical bug fixes.

---

## 📜 License

Kairo is released under the [MIT License](LICENSE).

---

## 🗺 Roadmap

- [ ] Multi-workspace support with encryption at rest
- [ ] Incremental DB-to-UI streaming for large datasets
- [ ] Conflict-free sync via append-only event log
- [ ] Sandboxed Plugin SDK
- [ ] Smart suggestions and spaced repetition
- [ ] Enhanced mobile/SSH terminal support
- [ ] Community plugin marketplace

---

## 💡 Philosophy

Kairo is built on the belief that task management should be **fast, simple, and under your control**. We prioritize:

✅ **Your Privacy** — Data stays on your machine  
✅ **Your Freedom** — Open source, MIT licensed  
✅ **Your Time** — Lightning-fast interactions  
✅ **Your Experience** — Premium, thoughtful design  

Every feature is carefully considered to maintain focus and avoid complexity creep.

---

## 📞 Support

- 🐛 Report bugs on [GitHub Issues](https://github.com/programmersd21/kairo/issues)
- 💬 Discuss ideas on [GitHub Discussions](https://github.com/programmersd21/kairo/discussions)
- ⭐ Show your support with a star!

---

**Made with ❤️ for focused execution. Start organizing your time today.**
