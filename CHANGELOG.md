# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
