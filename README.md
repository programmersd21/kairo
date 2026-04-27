<div align="center">

# 📚 Kairo

### ✨ A fast, keyboard-driven task manager that lives in your terminal — and never phones home.

![Demo](screenshots/demo.gif)

[![Release](https://img.shields.io/github/v/release/programmersd21/kairo?sort=semver\&style=for-the-badge\&logo=github\&color=7c3aed)](https://github.com/programmersd21/kairo/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/programmersd21/kairo/ci.yml?branch=main\&style=for-the-badge\&logo=githubactions\&logoColor=white\&color=2563eb)](https://github.com/programmersd21/kairo/actions)
[![Go Report Card](https://img.shields.io/badge/go%20report-A%2B-brightgreen?style=for-the-badge&logo=go&logoColor=white&color=10b981)](https://goreportcard.com/report/github.com/programmersd21/kairo)
[![License: MIT](https://img.shields.io/badge/License-MIT-f59e0b?style=for-the-badge)](https://opensource.org/licenses/MIT)

</div>

---

## ⚡ Try it in 10 seconds

```bash
go install github.com/programmersd21/kairo/cmd/kairo@latest
kairo
```

Press `n` → create your first task.

That’s it.

---

## 🧠 What is Kairo?

Kairo is a **terminal-first task system built for speed, automation, and control**.

No cloud.
No login.
No sync lock-in.
No background tracking.

Just your machine. Your workflow.

---

## 🔥 Why developers switch to Kairo

Most tools force trade-offs:

* GUI apps → slow, mouse-heavy, distracting
* plain text tools → no structure or querying
* cloud tools → lock-in + subscriptions
* old TUIs → functional but painful to use

Kairo exists in the gap:

> A **modern, scriptable, AI-aware task manager inside the terminal**

---

## ✨ Core experience

### 🔒 Local-first by design

* SQLite storage (WAL enabled)
* Fully offline
* Optional Git-based sync (no backend required)
* Export anytime: JSON / CSV / Markdown / text

### ⚡ Built for speed

* Sub-millisecond fuzzy search
* Full keyboard control (no mouse needed)
* Vim-style navigation (`j/k/gg/G`)
* Natural language deadlines:

  * `tomorrow 10am`
  * `next friday`
  * `in 2 hours`

### 🧩 Extensible like a toolchain

* Lua plugin system (event-driven)
* Headless CLI API for automation
* MCP server for AI agents
* Custom themes via code or config

### 🤖 AI built in (optional)

* Gemini integration (2.0 / 2.5 / 3.1 flash)
* Full task CRUD via chat
* Toggle anytime (`ctrl+a`)
* Never runs unless you invoke it

### 🎨 A terminal UI you won’t hate

* 32 built-in themes
* Live theme switching (`t`)
* Smooth animations for task actions
* Full-viewport rendering (clean UI, no clutter)

---

## 🧭 Feature snapshot

| Capability             | Status |
| ---------------------- | ------ |
| TUI with themes        | ✅      |
| Keyboard-only workflow | ✅      |
| Local-first storage    | ✅      |
| Git sync (no backend)  | ✅      |
| Lua plugins            | ✅      |
| CLI automation API     | ✅      |
| AI assistant           | ✅      |
| MCP server             | ✅      |

---

## 🚀 Key commands

```bash
kairo api create --title "Build project"
kairo api list --tag work
kairo api update --id <id> --status done
kairo sync
kairo export --format markdown
kairo mcp
```

---

## 🧠 Plugin system (Lua)

```lua
local plugin = {
    id = "notify-plugin",
    name = "Notifier"
}

kairo.on("task_create", function(event)
    kairo.notify("New task: " .. event.task.title)
end)

return plugin
```

---

## ⌨️ Keyboard shortcuts

| Key      | Action          |
| -------- | --------------- |
| `n`      | New task        |
| `e`      | Edit            |
| `z`      | Complete        |
| `d`      | Delete          |
| `t`      | Switch theme    |
| `f`      | Filter tags     |
| `ctrl+p` | Command palette |
| `ctrl+a` | AI panel        |
| `?`      | Help            |

---

## 🧱 Architecture

```
Input (CLI / UI / Lua / AI)
            ↓
      Task Engine (single source of truth)
            ↓
   SQLite (WAL) + Optional Git sync
            ↓
   Bubble Tea TUI (instant rendering)
```

---

## 🧭 Why Kairo exists

Most tools choose between:

* simplicity
* power
* speed

Kairo removes the trade-off.

---

## 📦 Full capability set

* Task views: Inbox, Today, Upcoming, Completed
* Tag-based filtering
* Fuzzy command palette
* Shell completions
* Import/export (JSON, CSV, Markdown)
* Git-backed sync
* MCP server for AI agents
* Plugin hooks for automation

---

## 🗺 Roadmap

* Encrypted multi-workspace support
* Event-sourced sync engine
* Plugin sandboxing
* Smart task suggestions
* Plugin marketplace

---

## 🤝 Contributing

PRs welcome.

Start with:

* themes
* plugins
* bug fixes
* documentation improvements

---

## ⭐ If this helps your workflow

Star the repo so more developers can find it.

---

<div align="center">

**Fast. Local. Scriptable. Yours.**

</div>
