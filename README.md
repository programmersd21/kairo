# Kairo

[![CI](https://github.com/programmersd21/kairo/actions/workflows/ci.yml/badge.svg)](https://github.com/programmersd21/kairo/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/programmersd21/kairo)](https://goreportcard.com/report/github.com/programmersd21/kairo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Time, executed well.**

Kairo is a keyboard-first, offline-first terminal task manager designed for focused execution. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and SQLite.

## ✨ Features

- **Task Engine:** Title, description (Markdown), tags, priority, deadline, status.
- **Views:** Inbox, Today, Upcoming, Tag, Priority.
- **Command Palette:** Ranked fuzzy search for tasks, commands, and tags.
- **Offline Storage:** SQLite with WAL + migrations for reliability and speed.
- **Git Sync:** Repo-backed, per-task JSON files, auto-commit, pull/push.
- **Plugins:** Lua-based commands and views with hot-reload.
- **Import/Export:** Support for JSON and Markdown.
- **Theming:** Built-in and user-definable theme overrides with runtime switching.

## 🏗 Architecture

Kairo is built with a modular architecture designed for performance, extensibility, and data sovereignty.

### 🧩 Core Components

- **UI Layer ([Bubble Tea](https://github.com/charmbracelet/bubbletea)):** An Elm-inspired functional TUI framework. Kairo uses a state-machine approach to manage different modes (List, Detail, Editor, Palette) and sub-component communication.
- **Storage Layer (SQLite):** A robust local database using `modernc.org/sqlite` (pure Go). It features WAL (Write-Ahead Logging) for concurrent access and a migration system for schema evolution.
- **Sync Engine (Git):** A unique "no-backend" synchronization strategy. It serializes tasks into individual JSON files within a local Git repository, leveraging Git's branching and merging capabilities for conflict resolution and versioning.
- **Search Engine:** An in-memory index utilizing a ranked fuzzy matching algorithm. It provides sub-millisecond search results by weighting matches based on contiguity and word boundaries.
- **Plugin System ([Gopher-Lua](https://github.com/yuin/gopher-lua)):** A lightweight Lua VM integration. It allows users to extend the TUI with custom commands and views without recompiling the binary.

### 🔄 Data Flow

1.  **Interaction:** User input is captured by the Bubble Tea loop and dispatched to the active component.
2.  **Persistence:** Changes are immediately persisted to the SQLite database.
3.  **Synchronization:** If enabled, the Sync Engine periodically (or on-demand) exports database state to the Git-backed task files and performs `git pull/push` operations.
4.  **Extensibility:** Lua plugins can hook into the task creation/deletion lifecycle and inject new items into the command palette.

## 🚀 Installation


```bash
kairo/
├── CHANGELOG.md
├── cmd
│   └── kairo
│       └── main.go
├── CODE_OF_CONDUCT.md
├── configs
│   └── kairo.example.toml
├── CONTRIBUTING.md
├── go.mod
├── go.sum
├── image.png
├── internal
│   ├── app
│   │   ├── model.go
│   │   └── msg.go
│   ├── config
│   │   └── config.go
│   ├── core
│   │   ├── codec
│   │   │   ├── json.go
│   │   │   └── markdown.go
│   │   ├── ids.go
│   │   ├── nlp
│   │   │   └── deadline.go
│   │   ├── task.go
│   │   └── view.go
│   ├── plugins
│   │   └── host.go
│   ├── search
│   │   ├── fuzzy.go
│   │   ├── fuzzy_test.go
│   │   └── index.go
│   ├── storage
│   │   ├── migrations.go
│   │   ├── repo.go
│   │   └── repo_test.go
│   ├── sync
│   │   └── engine.go
│   ├── ui
│   │   ├── detail
│   │   │   └── model.go
│   │   ├── editor
│   │   │   └── model.go
│   │   ├── help
│   │   │   └── model.go
│   │   ├── keymap
│   │   │   └── keymap.go
│   │   ├── palette
│   │   │   └── model.go
│   │   ├── styles
│   │   │   └── styles.go
│   │   ├── tasklist
│   │   │   └── model.go
│   │   ├── theme
│   │   │   └── theme.go
│   │   └── theme_menu
│   │       └── model.go
│   └── util
│       └── paths.go
├── LICENSE
├── Makefile
├── plugins
│   └── sample.lua
├── README.md
├── SECURITY.md
└── VERSION.txt
```

## 🚀 Installation

### Prerequisites

- Go **1.26+**

### Build from source

```bash
git clone https://github.com/programmersd21/kairo.git
cd kairo
make build
```

For a static binary (pure Go SQLite driver, no CGO):

```bash
CGO_ENABLED=0 make build
```

## 🛠 Usage

Run the binary:

```bash
./kairo
```

### Keybindings (Default)

- `ctrl+p`: Open command palette
- `n`: Create new task
- `e`: Edit selected task
- `d`: Delete selected task
- `enter`: View task details
- `1..5`: Switch views (Inbox, Today, Upcoming, Tag, Priority)
- `t`: Cycle theme
- `q` / `esc`: Back/Close

## ⚙️ Configuration

Copy the example configuration to your configuration directory:

- **Windows:** `%APPDATA%\kairo\config.toml`
- **macOS:** `~/Library/Application Support/kairo/config.toml`
- **Linux:** `~/.config/kairo/config.toml`

Example:
```bash
cp configs/kairo.example.toml ~/.config/kairo/config.toml
```

## 🔄 Git Sync

Enable sync in your `config.toml` and set `sync.repo_path` to a local git repository.

Kairo uses a distributed approach:
- Each task is stored as an individual JSON file.
- Changes are committed locally automatically.
- Manual sync: `kairo sync`

## 🔌 Plugins (Lua)

Kairo supports Lua plugins for custom commands and filters. Place `.lua` files in your `plugins/` directory.

Example: `plugins/sample.lua`

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for our code of conduct.

## 📜 License

Kairo is released under the [MIT License](LICENSE).

---

## 🗺 Roadmap

- [ ] Incremental DB-to-UI streaming for large datasets.
- [ ] Conflict-free sync via an append-only event log.
- [ ] Sandboxed Plugin SDK.
- [ ] Smart suggestions and spaced repetition.
- [ ] Multi-workspace support with encryption at rest.
