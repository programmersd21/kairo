# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.2]

### Added

* **Cinematic TUI Motion System**: A comprehensive motion engine for liquid glass interactions, elastic physics, and fluid boba clustering.
* **Cinematic View Shutter**: Smooth 600ms vertical split transition when switching tabs or closing menus, accompanied by a cascading task reveal.
* **Bulk Deletion API**: Added `kairo api delete all` to safely soft-delete all active tasks in one command.
* **Task Lifecycle Animations**: 
    * **Bloom**: New tasks expand into existence with an 800ms `EaseOutQuad` deliberate typing sequence.
    * **Glitch Deletion**: Bombastic 600ms glitch-vaporization effect where the task scrambles into particles and shrinks into nothingness.
    * **Liquid Fade**: Completed tasks "melt" into the background using progressive eased strikethrough.
* **Bento Layout System**: Redesigned header and empty states with modular, asymmetric blocks and soft, rounded borders for a premium aesthetic.

### Fixed

* **Isolated Tab Animations**: View transition bubbles in the header now *only* trigger when genuinely switching tabs, preventing layout flicker.
* **Context Isolation**: `Esc` now gracefully closes the Tag Filter, and selecting a plugin with `Enter` smoothly animates the view transition instead of instantly snapping.
* **Animation Glitches**: Resolved rendering artifacts (black blocks) during view transitions by ensuring background color persistence.
* **Layout Stability**: Fixed alignment of the empty state Bento card and footer keybindings.
* **Footer Legibility**: Fixed keybinding labels in the footer for better clarity.

## [1.2.1]

### Fixed

* **Footer Rendering**: Fixed an issue where the footer would disappear during theme changes or in specific application modes (Palette, Help, Theme Menu).
* **Footer Layout**: Optimized footer keybinding hints to be more compact, preventing layout wrapping and ensuring visibility on standard terminal widths.
* **UI Consistency**: Added missing icons to the footer across all modes for a more unified and polished user experience.
* **Component Sizing**: Fixed a layout bug where certain overlay modes (Palette, Help, Theme Menu) incorrectly occupied the full terminal height, obscuring the header and footer.

## [1.2.0]

### Added

* **Robust Tag Filter**: Real-time validation with instant feedback; prevents submission of invalid tags.
* **Multi-Tag Filtering**: Supports filtering by multiple tags (syntax: `work dev kairo`).
* **Continuous Database Cleanup**: Automatic hourly background pruning of orphaned data.
* **Database Cleanup API**: Added `kairo api cleanup` for manual pruning and vacuuming.
* **Premium Icon Overhaul**: Redesigned iconography using curated Nerd Font symbols.
* **Footer UX Enhancement**: Added descriptive icons for key actions; restored `f` shortcut.
* **Help Menu Redesign**: Cleaner structure with consistent iconography.
* **Safety First**: Quit confirmation dialog to prevent accidental exits.
* **Pill-Shaped Tags**: Introduced Powerline-style pill tag design.
* **Interactive Demo**: Added high-fidelity demo GIF to documentation.

### Changed

* **Enhanced Priority Badges**: Colored outlines matching priority levels.
* **Improved Delete Confirmation**: High-contrast red styling for clarity.
* **Pill-Shaped Tags**: Updated styling for better visual separation.
* **Tag Filtering System**: Migrated from single `tag` to multi `tags[]` across API, UI, Lua, and storage layers.

### Fixed

* **Priority Icon Mapping**: Fixed off-by-one display error.
* **Tag Filter Reset**: Empty `Enter` now correctly clears filter state.

## [1.1.9]

### Added

* **Premium Icon Overhaul**: A complete redesign of the application's iconography using a curated set of "Premium & Sentimental" symbols (Nerd Font optimized).
* **Footer UX Enhancement**: The footer now features descriptive icons for all key actions, including new dedicated symbols for GitHub Issues and Changelog access, and the restoration of the `f` tag filter shortcut.
* **Help Menu Redesign**: The help menu has been enhanced with professional icons for every keybinding category and action, providing a more intuitive and visually appealing reference.
* **Safety First**: Added a professional quit confirmation dialog to prevent accidental application closure.
* **Interactive Demo**: Added a new, high-fidelity demo GIF to the documentation for better visual clarity of the application's workflow and animations.

### Changed

* **Enhanced Priority Badges**: Task priority labels (P0-P3) now feature a colored outline matching the priority level for improved visual hierarchy and recognition.
* **Improved Delete Confirmation**: The delete confirmation badge in the footer now uses a high-contrast red background with white text for better visibility and safety.

### Fixed

* **Priority Icon Mapping**: Corrected an off-by-one error where priority icons were displaying $n+1$ for priority $n$.

## [1.1.8]

### Added

* **Startup Update Check**: Kairo now automatically checks for new releases on GitHub during startup. If a newer version is found, a notification appears in the footer with the version delta (e.g., `v1.1.7 → v1.1.8`) and instructions to update.
* **Smart Update Logic**: Update checks are intelligently skipped when running a development build (`dev`) to minimize noise for contributors.
* **Interactive Demo**: Added a new, high-fidelity demo GIF to the documentation for better visual clarity of the application's workflow and animations.

## [1.1.7]

### Added

* **GitHub Issues (`i`)**: Opens the GitHub issues page for the project in the default browser.
* **Changelog (`c`)**: Displays the `CHANGELOG.md` file within a dedicated TUI view.

## [1.1.6]

### Fixed

* **Windows Updater**: Resolved a critical issue where the binary update would fail due to file locking.

## [1.1.5]

### Added

* **New `help` Command**
* **Shell Tab Completions**
* **Completion Auto-Install**
* **Task ID in Detail View**
* **Editor Shortcut Toolbar**
* **Editor Clarity**
* **Multi-location Config Loading**

### Fixed

* **Rainbow Toggle**

### Changed

* **Active Tab Styling**

## [1.1.4]

### Changed

* **Linear Rainbow Logo Animation**

## [1.1.3]

### Added

* **Self-updating binary updater**
* **Cross-platform install scripts**
* **Plugin menu keybind footer**

### Changed

* Build metadata injection via GoReleaser

### Removed

* `go install` updater flow

## [1.1.2]

### Added

* **Plugin Metadata Display**
* **Uninstall Confirmation**

## [1.1.1]

### Added

* **20 New Themes (2026 Design Trends)**
* **Version Management Command**
* **Update Command**

### Fixed

* **.gorelease.yaml issue**

### Changed

* Theme registry expanded to 32 themes

## [1.1.0]

### Added

* Unified extensibility system
* Automation CLI API
* Enhanced Lua plugin system
* App lifecycle events
* Dynamic view shortcuts
* Tag filter keybinding

### Fixed

* Background rendering issues across UI

### Changed

* Refactored architecture to TaskService

## [1.0.4]

### Added

* Tag input modal overlay
* Explicit FilterState lifecycle
* Tag filter feedback in header

### Fixed

* Input focus conflicts
* Tag filter UI corruption
* Layout recalculation issues

### Changed

* Centralized input focus handling

## [1.0.0]

### Added

* Initial project release
* Core task engine
* Bubble Tea UI
* SQLite storage
* Git sync
* Lua plugins
* Documentation
