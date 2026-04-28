<div align="center">

# 📚 Kairo

### ✨ A fast, keyboard-driven task manager that lives in your terminal — and never phones home.

![Demo](screenshots/demo.gif)

[![Release](https://img.shields.io/github/v/release/programmersd21/kairo?style=for-the-badge\&logo=github\&color=7c3aed)](https://github.com/programmersd21/kairo/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/programmersd21/kairo/ci.yml?branch=main\&style=for-the-badge\&logo=githubactions\&color=2563eb)](https://github.com/programmersd21/kairo/actions)
[![Go Report Card](https://img.shields.io/badge/go%20report-A%2B-brightgreen?style=for-the-badge&logo=go&logoColor=white)](https://goreportcard.com/report/github.com/programmersd21/kairo)
[![License: MIT](https://img.shields.io/badge/License-MIT-f59e0b?style=for-the-badge)](https://opensource.org/licenses/MIT)

</div>

---

## ⚡ Install & run in seconds

Pick your platform:

### 🍺 macOS (Homebrew)

```bash
brew install programmersd21/kairo/kairo
```

### 🐧 Linux / macOS (curl)

```bash
curl -fsSL https://raw.githubusercontent.com/programmersd21/kairo/main/scripts/install.sh | bash
```

### 🪟 Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/programmersd21/kairo/main/scripts/install.ps1 | iex
```

### 🧰 Go install

```bash
go install github.com/programmersd21/kairo/cmd/kairo@latest
```

### ▶️ Run

```bash
kairo
```

Press `n` → create your first task.

Done writing? → `ctrl+s` to save your task!

---

## 🧠 One-line definition

Kairo is a **terminal task manager built in Go for developers who want speed, structure, and full local control**.

---

## 🔥 Why this exists

Most task managers are built wrong for developers:

* GUI apps → slow, mouse-driven, context switching
* Plain-text tools → flexible but no structure or querying
* Cloud apps → lock-in, subscriptions, data ownership loss
* Legacy TUIs → powerful but outdated UX

Kairo is the missing middle:

> A **modern, scriptable, AI-aware, local-first task system inside your terminal**

---

## ✨ Core capabilities

### 🔒 Data sovereignty

* SQLite storage (WAL-enabled)
* Fully offline operation
* Optional Git-backed sync (no backend)
* Export: JSON / CSV / Markdown / plain text

### ⚡ Speed at every layer

* Sub-millisecond fuzzy search with ranked results
* Full keyboard control (no mouse)
* Vim mode (j/k/gg/G)
* Natural language deadlines (tomorrow 10am, next friday, in 2 hours)

### 🧩 Extensibility

* Lua plugin system (event hooks: task_create, task_update, app_start, etc.)
* Headless CLI API for automation
* **Professional MCP Server**: Full Model Context Protocol implementation
    * **Tools**: CRUD operations, tag listing, and theme management
    * **Resources**: Direct access to JSON task data (`tasks://all`)
    * **Prompts**: Pre-configured AI workflows (`manage_tasks`)
* Custom themes via Lua or config

### 🤖 AI (optional)

* Gemini integration (2.0 / 2.5 / 3.1 flash)
* Full task CRUD from chat panel
* Toggle anytime (`ctrl+a`)
* Fully disabled unless invoked

### 🎨 Terminal UI

* Bento-style layout with Lip Gloss styling
* **Markdown Preview**: Side-by-side real-time rendering in the task editor (`ctrl+p`)
* 32 built-in themes (dark/light/hybrid)
* Live theme switching (`t`)
* Full-viewport rendering (no terminal bleed-through)
* Cinematic animations for create/complete/delete

---

## 🧭 Feature snapshot

| Capability             | Status |
| ---------------------- | ------ |
| Local-first storage    | ✅      |
| Full TUI with themes   | ✅      |
| Keyboard-only workflow | ✅      |
| Git sync (no backend)  | ✅      |
| Lua plugin system      | ✅      |
| CLI automation API     | ✅      |
| AI assistant           | ✅      |
| MCP server             | ✅      |
| Free & open source     | ✅      |

---

## 🚀 Quick commands

```bash
kairo api create --title "Finish report"
kairo api list --tag work
kairo api update --id <id> --status done
kairo export --format markdown
kairo sync
kairo mcp        # stdio mode (local)
kairo mcp 8080   # sse mode (remote)
```

---

## 🧱 Architecture

```
User Input (CLI / UI / Lua / AI)
        ↓
Task Service (single source of truth)
        ↓
SQLite (WAL) + optional Git sync
        ↓
Bubble Tea TUI (instant rendering)
```

---

## 🧠 Plugin system (Lua)

```lua
kairo.on("task_create", function(event)
    kairo.notify("New task: " .. event.task.title)
end)
```

Lua API: create_task, update_task, delete_task, list_tasks, on, notify

Events: task_create · task_update · task_delete · app_start · app_stop

---

## ⌨️ Keyboard shortcuts

| Key    | Action          |
| ------ | --------------- |
| n      | New task        |
| e      | Edit            |
| z      | Complete        |
| d      | Delete          |
| t      | Switch theme    |
| f      | Filter tags     |
| ctrl+p | Command palette |
| ctrl+a | AI panel        |
| ?      | Help            |

---

## 📦 Full capability set

* Views: Inbox, Today, Upcoming, Completed, by Tag, Priority
* Fuzzy command palette (ctrl+p)
* Shell completions (bash/zsh/fish/powershell)
* Import/export: JSON, CSV, Markdown
* Git-backed sync (per-task JSON)
* MCP server for AI agents
* Plugin hooks for automation

---

## 🚀 Automation API

```bash
kairo api create --title "task" --priority 1
kairo api list --tag work
kairo api update --id <id> --status done
kairo api delete all
```

JSON mode:

```bash
kairo api --json '{"action":"create","payload":{"title":"API task"}}'
```

Extras:

```bash
kairo api set_theme --theme nord
kairo api plugin_list
kairo export --format csv
kairo sync
kairo mcp
```

---

## 🧩 Plugin system (full example)

```lua
local plugin = {
    id = "my-plugin",
    name = "My Plugin",
    version = "1.0.0"
}

kairo.on("task_create", function(event)
    kairo.notify("New task: " .. event.task.title)
end)

return plugin
```

---

## 🧱 Architecture stack

* TUI: Bubble Tea (state machine)
* Styling: Lip Gloss
* Storage: SQLite (WAL)
* Search: in-memory fuzzy index
* Plugins: GopherLua VM
* Sync: Git-based per-task files
* AI: Gemini tool-calling API

---

## 🗺 Roadmap

* encrypted multi-workspace support
* event-sourced sync engine
* sandboxed plugins
* smart task suggestions
* plugin marketplace
* streaming performance optimizations

---

## ⚙️ Configuration

Auto-generated on first run:

* Linux: `~/.config/kairo/config.toml`
* macOS: `~/Library/Application Support/kairo/config.toml`
* Windows: `%APPDATA%\\kairo\\config.toml`

In-app config: `ctrl+s` (settings menu)

---

## 🤝 Contributing

PRs welcome:

* themes
* plugins
* performance
* docs

Special thanks to @Tornado300 for key bug fixes and contributions.

---

## ⭐ Star this repo

If Kairo improves your workflow, star it so more developers can discover it.

---

<div align="center">

**Fast. Local. Scriptable. Terminal-native.**

</div>
