<div align="center">

# Kairo

### A keyboard-driven task manager that lives in your terminal and never phones home.

![Demo](screenshots/demo.gif)

[![Release](https://img.shields.io/github/v/release/programmersd21/kairo?sort=semver&style=for-the-badge&logo=github&color=7c3aed)](https://github.com/programmersd21/kairo/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/programmersd21/kairo/ci.yml?branch=main&style=for-the-badge&logo=githubactions&logoColor=white&color=2563eb)](https://github.com/programmersd21/kairo/actions)
[![Go Report Card](https://img.shields.io/badge/go%20report-A%2B-brightgreen?style=for-the-badge&logo=go&logoColor=white&color=10b981)](https://goreportcard.com/report/github.com/programmersd21/kairo)
[![License: MIT](https://img.shields.io/badge/License-MIT-f59e0b?style=for-the-badge&logo=open-source-initiative&logoColor=white)](https://opensource.org/licenses/MIT)

</div>

---

## Why this exists

Most task managers are web apps with offline modes bolted on. The rest are plain-text systems that break the moment you need structure.

- **GUI apps** pull you out of flow and require a mouse
- **Plain-text tools** have no querying, filtering, or automation surface
- **Cloud-sync apps** own your data — and charge you for access to it
- **Existing TUI tools** are functional but aesthetically hostile — they look like it's 1992
- **Nothing** gives you a scriptable, themeable, AI-aware task manager that runs entirely on your machine

Kairo is the gap between those two worlds.

---

## What it is

Kairo is a terminal task manager built in Go. It gives you a full TUI with smooth animations and 32 themes, a headless CLI API for scripting, a Lua plugin system for custom logic, an optional Git-backed sync engine, and a built-in Gemini AI assistant — all storing data locally in SQLite. No account. No server. No subscription.

---

## Core values

### 🔒 Data sovereignty
- SQLite on disk, WAL-enabled for concurrent access
- Optional Git sync: per-task JSON files, no backend, no lock-in
- Export any time to JSON, CSV, Markdown, or plain text

### ⌨️ Speed at every layer
- Sub-millisecond fuzzy search with ranked results
- Full keyboard control — mouse never required
- Vim mode available for home-row navigation (`j`/`k`/`gg`/`G`)
- Natural language deadlines: `tomorrow at 10am`, `next friday`, `in 2 hours`

### 🧩 Extensibility
- Lua plugins with a full event hook system (`task_create`, `task_update`, `app_start`, etc.)
- Stable headless CLI API — scriptable from any shell or CI pipeline
- Built-in MCP server exposing the full task schema to AI agents
- Custom themes definable via Lua or the API

### 🤖 AI that stays optional
- Integrated Gemini assistant (2.0/2.5/3.1 flash) with full task control from the chat panel
- Configure with one command, disable just as easily
- AI never runs unless you invoke it

### 🎨 A terminal UI you'll actually want to open
- Bento-style layout with soft, rounded components (built on [Lip Gloss](https://github.com/charmbracelet/lipgloss))
- 32 built-in themes — dark, light, and hybrid — switchable live with `t`
- Full-viewport background rendering: no terminal bleed-through
- Cinematic animations on create, complete, and delete

---

## Quick start

**macOS (Homebrew):**
```bash
brew tap programmersd21/kairo_tap && brew install --cask kairo
```

**Linux / macOS (curl):**
```bash
curl -fsSL https://raw.githubusercontent.com/programmersd21/kairo/main/scripts/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/programmersd21/kairo/main/scripts/install.ps1 | iex
```

**Go:**
```bash
go install github.com/programmersd21/kairo/cmd/kairo@latest
```

Then launch:
```bash
kairo
```

That's it. Press `n` to create your first task. Press `?` for help.

> Kairo checks for updates on startup. Run `kairo update` to upgrade in place — binary is verified against `checksums.txt` automatically.

---

## See it in action

> 📽️ _Full demo GIF — [`screenshots/demo.gif`](screenshots/demo.gif)_

<!-- Replace with actual GIF or screenshot grid showing: task list, command palette, AI panel, theme switcher -->

---

## How it compares

|  | Kairo | Taskwarrior | Todoist / Linear | plain `.txt` |
|---|---|---|---|---|
| Full TUI with themes | ✅ | ❌ | ❌ | ❌ |
| Keyboard-only control | ✅ | ✅ | ❌ | ✅ |
| Local-first storage | ✅ | ✅ | ❌ | ✅ |
| Git sync (no backend) | ✅ | ❌ | ❌ | manual |
| Lua plugin system | ✅ | ❌ | ❌ | ❌ |
| Headless CLI API | ✅ | partial | ❌ | ❌ |
| AI assistant | ✅ | ❌ | partial | ❌ |
| MCP server | ✅ | ❌ | ❌ | ❌ |
| Free / open source | ✅ | ✅ | ❌ | ✅ |

Kairo is not trying to replace project management software. It is a fast, local-first personal task layer that you can automate, extend, and trust.

---

## Full feature reference

**Task management**
- Views: Inbox, Today, Upcoming, Completed, by Tag, by Priority
- Natural language deadline parsing (`today`, `next friday`, `in 2 hours`, `august 24`)
- Fuzzy command palette (`ctrl+p`) with ranked results
- Tag-based filtering with multi-tag support

**UI & navigation**
- 32 built-in themes, live-switchable with `t`; custom themes via Lua
- Vim mode (`j`/`k`/`gg`/`G`) — opt-in, toggle in settings
- Responsive layout — adapts gracefully to any terminal size
- Shell completions for bash, zsh, fish, PowerShell

**Automation & integration**
- Headless CLI API: `kairo api create/list/update/delete` with JSON interface
- Import/export: JSON, CSV, Markdown, plain text
- Git-backed sync: per-task JSON files, committed locally, synced on demand
- Built-in MCP server (`kairo mcp`) for AI agent access to full task schema

**AI assistant**
- Gemini integration (2.0 / 2.5 / 3.1 flash, switchable live with ←/→)
- Full task CRUD from the chat panel
- Toggle with `ctrl+a`; clear history with `ctrl+l`

**Plugins (Lua)**
- Event hooks: `task_create`, `task_update`, `task_delete`, `app_start`, `app_stop`
- Custom commands, themes, and UI extensions
- Full Lua API: `kairo.create_task()`, `kairo.list_tasks()`, `kairo.notify()`, and more

---

## Automation API

Every TUI operation is available headlessly:

```bash
# Task operations
kairo api list --tag work
kairo api create --title "Finish report" --priority 1
kairo api update --id <id> --status done
kairo api delete all

# JSON interface
kairo api --json '{"action": "create", "payload": {"title": "API task", "tags": ["bot"]}}'

# Theme and plugin management
kairo api set_theme --theme nord
kairo api plugin_list
kairo api plugin_get --name auto-cleanup.lua

# Configure AI
kairo api configure-ai set "YOUR_GEMINI_API_KEY"

# Export
kairo export --format csv --out tasks.csv
kairo export --format markdown --out tasks.md

# Sync
kairo sync

# Start MCP server
kairo mcp
```

---

## Plugin system (Lua)

```lua
-- plugins/my-plugin.lua
local plugin = {
    id = "my-plugin",
    name = "My Plugin",
    version = "1.0.0",
}

kairo.on("task_create", function(event)
    kairo.notify("New task: " .. event.task.title)
end)

plugin.commands = {
    { id = "hello", title = "Say Hello", run = function() kairo.notify("Hello!") end }
}

return plugin
```

**Lua API surface:** `create_task`, `update_task`, `delete_task`, `list_tasks`, `on`, `notify`

**Events:** `task_create` · `task_update` · `task_delete` · `app_start` · `app_stop`

---

## Keyboard reference

| Key | Action |
|-----|--------|
| `ctrl+p` | Command palette |
| `n` | New task |
| `e` | Edit task |
| `z` | Toggle complete (animated) |
| `d` | Delete task |
| `t` | Cycle themes |
| `f` | Tag filter |
| `1`–`9` | Switch views |
| `ctrl+a` | AI assistant panel |
| `ctrl+s` | Settings |
| `?` | Help |
| `q` | Quit |

---

## Architecture

```
User Input / CLI API / Lua Plugin
          ↓
      Task Service  ←──── single source of truth
          ↓
  SQLite (WAL)  +  Optional Git Sync
          ↓
   Bubble Tea UI  →  instant re-render
```

| Layer | Technology |
|-------|-----------|
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) — Elm-inspired, state-machine driven |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Storage | SQLite with WAL (pure Go) |
| Search | In-memory fuzzy index, sub-millisecond |
| Plugins | [GopherLua](https://github.com/yuin/gopher-lua) — embedded Lua VM |
| Sync | Git, per-task JSON files |
| AI | Gemini API with tool-calling |

---

## Roadmap

- [ ] Multi-workspace support with encryption at rest
- [ ] Conflict-free sync via append-only event log
- [ ] Sandboxed plugin SDK with capability permissions
- [ ] Smart task suggestions and spaced repetition
- [ ] Community plugin marketplace
- [ ] Incremental streaming for large datasets

---

## Configuration

Config is auto-created on first run. To customize manually:

| OS | Path |
|----|------|
| Linux | `~/.config/kairo/config.toml` |
| macOS | `~/Library/Application Support/kairo/config.toml` |
| Windows | `%APPDATA%\kairo\config.toml` |

Open the in-app settings menu with `ctrl+s` — no file editing required for most options.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Good entry points:

- New themes (Lua or config)
- Bug fixes and performance improvements
- Plugin examples and documentation
- Translations

Special thanks to **@Tornado300** for surfacing several critical bug fixes.

---

## License

MIT — [LICENSE](LICENSE)

---

<div align="center">

**Your tasks. Your machine. Your rules.**

[Report a bug](https://github.com/programmersd21/kairo/issues) · [Start a discussion](https://github.com/programmersd21/kairo/discussions) · [⭐ Star on GitHub](https://github.com/programmersd21/kairo)

</div>
