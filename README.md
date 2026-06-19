<div align="center">

<img src="screenshots/logo.png" alt="Kairo" width="80" />

# Kairo

**The terminal task manager for developers who live in their editor.**

Kairo is a keyboard-first task manager for the terminal. It keeps tasks, projects, recurring work, and notes local, fast, and easy to automate.

<br/>

[![Release](https://img.shields.io/github/v/release/programmersd21/kairo?style=for-the-badge\&logo=github\&color=7c3aed)](https://github.com/programmersd21/kairo/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/programmersd21/kairo/ci.yml?branch=main\&style=for-the-badge\&logo=github\&color=2563eb)](https://github.com/programmersd21/kairo/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-f59e0b?style=for-the-badge)](https://opensource.org/licenses/MIT)

<br/>

![Kairo Demo](screenshots/demo.gif)

</div>

**YouTube tutorial playlist (in dev):**
[https://youtube.com/playlist?list=PLvaz_NYJcySmNh28QzxLqV5HRrTslEaUo&si=XY8YvRnrhqxqU6RD](https://youtube.com/playlist?list=PLvaz_NYJcySmNh28QzxLqV5HRrTslEaUo&si=XY8YvRnrhqxqU6RD)

---

## Why Kairo?

Kairo is built for people who want speed, structure, and full keyboard control without leaving the terminal.

| Feature                  | Benefit                                           |
| ------------------------ | ------------------------------------------------- |
| **Momentum Dashboard**   | Quick view of task progress and activity.         |
| **White-Space-First UI** | Focused layout with less visual noise.            |
| **Monochrome Design**    | Clean core with semantic color accents.           |
| **Typography Hierarchy** | Clear visual focus through font weight and scale. |
| **Fluid Motion**         | Fast transitions that stay out of the way.        |

---

## Quick Start

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

Then run:

```bash
kairo
```

Press `n` to create your first task. Press `ctrl+s` to open settings.

> Works best on Alacritty. Some terminals may have rendering quirks - see [#16](https://github.com/programmersd21/kairo/issues/16).

---

## Features

![Kairo Home Screen](screenshots/home_screen.png)

### Genuinely Fast

Sub-millisecond fuzzy search, Vim bindings (`j/k/gg/G`), and full keyboard control.

### Project Sidebar & Hierarchy

Organize work into projects and nested task trees. Toggle the sidebar with `ctrl+e` and switch projects with arrow keys.

### Recurring Tasks & Results

Create weekly or monthly schedules, and attach completion notes when marking work as done.

### Your Data, Locally

SQLite with WAL mode. Fully offline. Optional Git-backed sync. Export to JSON, CSV, Markdown, or plain text.

### Interactive Stats Dashboard & Focus Engine

Press `s` to open statistics and `f` to launch the built-in Pomodoro timer.

### Optional AI

Gemini integration is available for natural-language task creation and management. Toggle it with `ctrl+a`.

### Beautiful by Default

32 built-in themes, live switching, Markdown preview, and configurable UI behavior.

### Extensible to the Core

Lua plugins, a headless CLI API, and an MCP server for agent workflows.

### Undo & Redo

Track task creation, editing, deletion, and status changes with local history.

### Tag Highlighting

Color-code tags directly in `config.toml`.

```toml
[tags.highlight]
work    = { fg = "#CCCCCC" }
private = "fg=#EEEEEE,bg=#0000FF,bold"
diy     = "bg=accent"
```

---

## Keyboard Shortcuts

| Key      | Action                             |
| -------- | ---------------------------------- |
| `n`      | New task                           |
| `D`      | Duplicate task                     |
| `e`      | Edit task                          |
| `z`      | Complete task                      |
| `ctrl+d` | Duplicate task                     |
| `Space`  | Select task / Collapse subtasks    |
| `s`      | Stats dashboard                    |
| `f`      | Focus engine                       |
| `ctrl+f` | Filter by tag                      |
| `ctrl+e` | Switch project                     |
| `p`      | Manage plugins                     |
| `t`      | Switch theme                       |
| `ctrl+p` | Command palette / Markdown preview |
| `ctrl+a` | AI panel                           |
| `ctrl+s` | Settings                           |
| `x`      | Import / Export                    |
| `?`      | Help                               |
| `ctrl+z` | Undo last action                   |
| `ctrl+y` | Redo last undone                   |
| `ctrl+w` | Welcome tour                       |

<div align="center">
  <img src="screenshots/new_task.png" width="30%" />
  <img src="screenshots/filter_tags.png" width="30%" />
  <img src="screenshots/help_menu.png" width="30%" />
  <img src="screenshots/settings_menu.png" width="30%" />
  <img src="screenshots/theme_menu.png" width="30%" />
  <img src="screenshots/dashboard.png" width="30%" />
  <img src="screenshots/welcome_tour.png" width="30%" />
  <img src="screenshots/focus_mode.png" width="30%" />
  <img src="screenshots/palette.png" width="30%" />
</div>

---

## CLI Automation

Kairo exposes a full CLI API for scripting and CI pipelines, with support for `parent_id` and `collapsed` state:

```bash
# Create a task
kairo api create --title "Finish report" --priority 1

# List by tag
kairo api list --tag work

# Mark complete
kairo api update --id <id> --status done

# Export everything
kairo export --format markdown
```

---

## Lua Plugin System

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

Browse [sample plugins →](https://github.com/programmersd21/kairo/tree/main/plugins)

---

## Architecture

```text
Input  (CLI · TUI · Lua · AI)
       ↓
Task Service  (single source of truth)
       ↓
SQLite (WAL)  +  optional Git sync
       ↓
Bubble Tea TUI  (instant rendering)
```

**Stack:** Bubble Tea · Lip Gloss · SQLite (WAL) · GopherLua · Gemini API · Git

---

## Everything Included

| Feature                    | Status |
| -------------------------- | ------ |
| Local-first SQLite storage | ✅      |
| Nested tasks & folders     | ✅      |
| 32 themes, live switching  | ✅      |
| Keyboard-only workflow     | ✅      |
| Recurring tasks            | ✅      |
| Git sync (no backend)      | ✅      |
| Lua plugin system          | ✅      |
| CLI automation API         | ✅      |
| AI assistant (optional)    | ✅      |
| MCP server                 | ✅      |
| Free & open source         | ✅      |

---

## Configuration

Kairo can be configured via `config.toml` in your application data directory.

### Task List

You can customize the fields shown on the right side of the task list:

```toml
[list.order]
right = ["tags", "due", "priority"]
```

Valid values for `right` are: `tags`, `due`, `priority`.

### Task Fields

**Minimal Due Mode**: Abbreviate overdue states and use a fixed-width column for consistent alignment.

```toml
[list.fields.due]
minimal = true
```

**wait_until**: Hide a task from the task list until the specified datetime.

**until**: Stop generating new recurring instances after the specified datetime.

Auto-generated on first run at:

* **Linux:** `~/.config/kairo/config.toml`
* **macOS:** `~/Library/Application Support/kairo/config.toml`
* **Windows:** `%APPDATA%\kairo\config.toml`

| Option       | Description             | Default      |
| ------------ | ----------------------- | ------------ |
| `theme`      | UI theme name           | `catppuccin` |
| `vim_mode`   | Vim keybindings         | `false`      |
| `show_help`  | Help footer             | `true`       |
| `show_id`    | Task IDs in detail view | `true`       |
| `animations` | UI animations           | `true`       |
| `rainbow`    | Animated rainbow logo   | `false`      |

Prefer not to edit files? `ctrl+s` opens the in-app settings menu.

---

## Roadmap

* Encrypted multi-workspace support
* Event-sourced sync engine
* Sandboxed plugin environment
* Smart task suggestions
* Plugin marketplace
* Streaming performance optimizations

---

## Star History

<a href="https://www.star-history.com/?repos=programmersd21%2Fkairo&type=date&legend=top-left">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=programmersd21/kairo&type=date&theme=dark&legend=top-left" />
    <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=programmersd21/kairo&type=date&legend=top-left" />
    <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=programmersd21/kairo&type=date&legend=top-left" />
  </picture>
</a>

---

## Contributing

PRs are welcome - especially for themes, plugins, performance, and docs. If something bugs you, fix it.

Huge thanks to [@Tornado300](https://github.com/Tornado300), [@riodelphino](https://github.com/riodelphino) and [@FuryRacer](https://github.com/FuryRacer) for key bug fixes and improvements that made Kairo better for everyone.

---

<div align="center">

**If Kairo saves you time, a ⭐ helps other developers find it.**

<br/>

*Built for the terminal. Built for focus. Built for you.*

</div>
