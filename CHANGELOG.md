# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.4]

### Changed
- **Linear Rainbow Logo Animation**: KAIRO logo now animates with a smooth, linear rainbow color shift.

## [1.1.3]

### Added
- **Self-updating binary updater**: `kairo update` now downloads the correct GitHub Release asset for your OS/arch, verifies it against `checksums.txt`, and performs a safe in-place binary swap (with `.old` backup/rollback).
- **Cross-platform install scripts**: `scripts/install.sh` (Linux/macOS) and `scripts/install.ps1` (Windows) install into standard user locations and add the install directory to PATH when possible.
- **Plugin menu keybind footer**: plugin manager overlay now shows a quick keybind legend (`enter`, `u`, `esc`, etc.).

### Changed
- `kairo version` now prints build version + commit (when available).
- GoReleaser now injects build metadata into `internal/buildinfo` (instead of `main.*`).

### Removed
- `go install`-based updater flow (replaced by the GitHub Releases updater).

## [1.1.2]

### Added
- **Plugin Metadata Display**: Press `Enter` on a plugin in the menu to view full metadata including Name, Description, Author, and Version.
- **Uninstall Confirmation**: Added safety confirmation dialog before uninstalling plugins with `u` key.

## [1.1.1]

### Added
- **20 New Themes (2026 Design Trends)**:
  - Dark themes: `obsidian_bloom`, `neon_reef`, `carbon_sunset`, `vanta_aurora`, `plasma_grape`, `midnight_jade`, `synthwave_minimal`, `graphite_matcha`
  - Light themes: `cloud_dancer`, `sakura_sand`, `olive_mist`, `terracotta_air`, `vanilla_sky`, `peach_fuzz_neo`, `coastal_drift`, `matcha_latte`
  - Hybrid themes: `digital_lavender`, `neo_mint_system`, `sunset_gradient_pro`, `forest_sanctuary`
- **Version Management**: `kairo version` command to display installed version
- **Update Command**: `kairo update` command for one-step updates via `go install github.com/programmersd21/kairo/cmd/kairo@latest`

### Fixed
- **.gorelease.yaml** was failing on homebrew step, so it was resolved.

### Changed
- Updated theme registry to 32 total themes (12 legacy + 20 new)

## [1.1.0]

### Added
- **Unified Extensibility System**: A shared task service layer for TUI, Lua, and CLI.
- **Automation CLI API**: Stable `kairo api` command for external scripting and JSON integration.
- **Enhanced Lua Plugin System**: 
    - Full Task CRUD access via `kairo` module.
    - Event Hook System subscribing to `task_create`, `task_update`, `task_delete`, `app_start`, and `app_stop`.
    - Improved Plugin Host with robust error handling and unified engine.
- **App Lifecycle Events**: Proper emission of start/stop events for plugin orchestration.
- **Dynamic View Shortcuts**: `1-9` keys now switch to the corresponding tab index, working for all built-in and plugin-provided views.
- **Specific Tag Filter Key**: `f` now specifically switches to the Tag View and opens the filter input modal.

### Fixed
- **Background Rendering Bleed-Through**: Resolved a visual bug where the terminal's default background color showed through in whitespace regions, creating inconsistent visuals across the entire viewport.
  - **Root Cause**: Inline spacer strings (`strings.Repeat(" ", N)`) in the header, footer, and task rows were plain text without ANSI background escape codes. Additionally, multiple Lip Gloss styles (`Muted`, `Accent`, `Badge*`, `Tab*`) were defined without `.Background()`, causing their ANSI reset codes to clear the container's background.
  - Added explicit `.Background(t.Bg)` to all content-level styles (`Muted`, `Accent`, `TabActive`, `TabInactive`, `Badge`, `BadgeGood`, `BadgeWarn`, `BadgeBad`, `BadgeMuted`).
  - Wrapped all inline spacer strings in styled renders with the theme background color.
  - Added background to the detail view outer container.
  - The fix is robust across resizing, scrolling, theme switching, and all UI modes.
### Changed
- Refactored internal architecture to use `TaskService` as the single source of truth.
- Standardized Lua plugin structure with metadata, commands, and views.
- Improved CLI consistency with new `api` subcommand flags and JSON support.

## [1.0.4]

### Added
- Direct tag input via keyboard modal overlay - press `4` in Tag View to open tag filter input
- Explicit FilterState lifecycle management for robust filter state handling
- Tag filter visual feedback with active filter indicator in header

### Fixed
- Global keybindings no longer trigger while typing in input fields (palette, editor, tag filter)
- Tag filter UI rendering corruption - tabs no longer disappear when filter is applied
- Tag filter state management - filter can now be properly cleared and edited
- Layout recalculation on filter state changes prevents component overflow

### Changed
- Input focus protection centralized via `isInputFocused()` helper for consistent behavior across all input modes
- rebuildComponentSizes() now called automatically when filter state changes

## [1.0.0]

### Added
- Initiated the project.
- Project scaffolding and initial core logic.
- Task engine with title, description, tags, priority, and deadline.
- Bubble Tea UI with multiple views (Inbox, Today, Upcoming, etc.).
- SQLite storage with migrations.
- Git-backed sync engine.
- Lua plugin support.
- Initial project documentation and repository structure.
