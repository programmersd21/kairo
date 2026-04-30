<div align="center">

# 📚 Kairo

**A terminal task manager built for developers who are tired of fighting their tools.**

Fast. Local. Yours.

![Demo](screenshots/demo.gif)

<br/>

[![Release](https://img.shields.io/github/v/release/programmersd21/kairo?style=for-the-badge&logo=github&color=7c3aed)](https://github.com/programmersd21/kairo/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/programmersd21/kairo/ci.yml?branch=main&style=for-the-badge&logo=githubactions&color=2563eb)](https://github.com/programmersd21/kairo/actions)
[![Go Report Card](https://img.shields.io/badge/go%20report-A%2B-brightgreen?style=for-the-badge&logo=go&logoColor=white)](https://goreportcard.com/report/github.com/programmersd21/kairo)
[![License: MIT](https://img.shields.io/badge/License-MIT-f59e0b?style=for-the-badge)](https://opensource.org/licenses/MIT)

</div>

---

You know that feeling when your task manager gets in the way of your actual work?

Kairo was built because of that feeling.

No browser tabs. No subscriptions. No mouse. Just you, your terminal, and your tasks — right where your brain already lives.

---

## ⚡ Get started in seconds

**macOS (Homebrew)**
```bash
brew install programmersd21/kairo/kairo
```

**Linux / macOS**
```bash
curl -fsSL https://raw.githubusercontent.com/programmersd21/kairo/main/scripts/install.sh | bash
```

**Windows (PowerShell)**
```powershell
iwr -useb https://raw.githubusercontent.com/programmersd21/kairo/main/scripts/install.ps1 | iex
```

**Go**
```bash
go install github.com/programmersd21/kairo/cmd/kairo@latest
```

Then just run:
```bash
kairo
```

Press `n` to create your first task. `ctrl+s` to save it. That's it.

> Works best on Alacritty. Some terminals may have rendering quirks — see [#16](https://github.com/programmersd21/kairo/issues/16).

---

## 💫 Star History

<a href="https://www.star-history.com/?repos=programmersd21%2Fkairo&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=programmersd21/kairo&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=programmersd21/kairo&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=programmersd21/kairo&type=date&legend=top-left" />
 </picture>
</a>

---

## 🧠 Why Kairo?

Most tools ask you to adapt to them. Kairo adapts to you.

| The problem | What Kairo does instead |
|---|---|
| GUI apps pull you away from your flow | Lives entirely in your terminal |
| Cloud tools own your data | Everything stays local, in SQLite |
| Plain-text tools lack structure | Full tagging, filtering, and fuzzy search |
| Legacy TUIs feel clunky | Modern, animated, keyboard-first UX |

Your tasks are yours. They don't belong in someone else's cloud.

---

## ✨ What it can do

### It's fast — genuinely fast
Sub-millisecond fuzzy search. Full keyboard control. Vim bindings (`j/k/gg/G`). Natural language deadlines like `tomorrow 10am` or `next friday`. You never have to leave the keyboard.

### It respects your data
SQLite storage with WAL mode. Fully offline. Optional Git-backed sync — no backend, no account, no lock-in. Export to JSON, CSV, Markdown, or plain text whenever you want.

### It grows with you
A Lua plugin system lets you hook into task events. A headless CLI API means you can automate anything. And an MCP server opens Kairo up to AI agents that can read and manage your tasks directly.

### AI — when you want it, invisible when you don't
Optional Gemini integration (2.0 / 2.5 / 2.5 Flash). Toggle it with `ctrl+a`. It never runs unless you invoke it. Your workflow, your call.

### Beautiful by default
32 built-in themes. Live switching with `t`. Bento-style layout. Real-time Markdown preview (`ctrl+p`). Cinematic animations for create, complete, and delete (with a global toggle in `ctrl+s` to disable them for maximum speed). It's a terminal app that you'll actually enjoy looking at.

---

## ⌨️ The shortcuts that matter

| Key | What it does |
|---|---|
| `n` | New task |
| `e` | Edit |
| `z` | Complete |
| `d` | Delete |
| `t` | Switch theme |
| `f` | Filter by tag |
| `ctrl+p` | Command palette |
| `ctrl+a` | AI panel |
| `?` | Help |
| `ctrl+s` | Settings |
| `x` | Import/Export |

---

## 🚀 Automate everything

Kairo has a full CLI API for scripting and pipelines:

```bash
# Create a task from anywhere
kairo api create --title "Finish report" --priority 1

# List tasks by tag
kairo api list --tag work

# Mark done
kairo api update --id <id> --status done

# Export your whole list
kairo export --format markdown

# Sync via Git
kairo sync

# Start the MCP server (for AI agents)
kairo mcp        # stdio mode
kairo mcp 8080   # SSE mode
```

JSON mode for maximum scriptability:
```bash
kairo api --json '{"action":"create","payload":{"title":"API task"}}'
```

---

## 🧩 Extend it with Lua

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

Available hooks: `task_create` · `task_update` · `task_delete` · `app_start` · `app_stop`

Available API: `create_task`, `update_task`, `delete_task`, `list_tasks`, `on`, `notify`

**Find sample plugins [here](https://github.com/programmersd21/kairo/tree/main/plugins).**

---

## 🧱 How it's built

```
Your input (CLI / TUI / Lua / AI)
          ↓
  Task Service (single source of truth)
          ↓
  SQLite (WAL) + optional Git sync
          ↓
  Bubble Tea TUI (instant rendering)
```

**Stack:** Bubble Tea · Lip Gloss · SQLite (WAL) · GopherLua · Gemini API · Git

---

## ✅ Everything included

| Feature | Status |
|---|---|
| Local-first SQLite storage | ✅ |
| Full TUI with 32 themes | ✅ |
| Keyboard-only workflow | ✅ |
| Git sync (no backend) | ✅ |
| Lua plugin system | ✅ |
| CLI automation API | ✅ |
| AI assistant (optional) | ✅ |
| MCP server | ✅ |
| Free & open source | ✅ |

---

## ⚙️ Configuration

Auto-generated on first run:

- **Linux:** `~/.config/kairo/config.toml`
- **macOS:** `~/Library/Application Support/kairo/config.toml`
- **Windows:** `%APPDATA%\kairo\config.toml`

Prefer in-app settings? `ctrl+s` opens the settings menu.

---

## 🗺 What's coming

- Encrypted multi-workspace support
- Event-sourced sync engine
- Sandboxed plugin environment
- Smart task suggestions
- Plugin marketplace
- Streaming performance optimizations

---

## 🤝 Contributing

If something bugs you, fix it. PRs are welcome — especially for themes, plugins, performance, and docs.

A huge thank you to [@Tornado300](https://github.com/Tornado300) for key bug fixes and contributions that made Kairo better for everyone.

---

## ⭐ If Kairo helps you

Star the repo. It's the simplest way to help other developers find this tool and spend less time fighting their workflow.

---

<div align="center">

*Built for the terminal. Built for focus. Built for you.*

</div>
