# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.5]
- **CLI Validation**: Added robust validation for subcommands and flags. Kairo now warns the user and provides helpful guidance when an invalid command or flag is provided.
- **Global Flags**: Added support for `-h`/`--help` and `-v`/`--version` as global flags.
- **Enhanced Error Feedback**: Improved CLI error messages to include specific details about unknown arguments and automatically list available commands or flags.

## [1.3.4]

* **Editor Preview Hint**: Added a visual keybinding hint (`ctrl+p preview`) to the task editor footer. This makes the real-time Markdown preview feature more discoverable for new users.


## [1.3.3]

* **Global Animations Toggle**: Added a new "Animations" setting (default: on) to toggle app-wide cinematic effects. When disabled, the UI bypasses all creation, completion, deletion, and view transition animations for a snappier, instant-feedback experience.
* **Settings Navigation**: Added support for `j` (move down) and `k` (move up) in the Settings menu for improved keyboard accessibility.
* **Shortcuts Update**: Standardized keybindings for core utilities: `ctrl+s` for Settings and `x` for Import/Export.

## [1.3.2]

### Added

* **MCP Subcommand**: Properly registered the `mcp` subcommand in the CLI help output.
* **MCP Startup Logs**: Enhanced the MCP server startup logs to display the active listening address and port.

## [1.3.1]

### Added

* **Markdown Preview Panel**: The task editor now features a side-by-side markdown preview panel (toggled with `ctrl+p` or automatically on wide screens) for real-time visualization of task descriptions.
* **Plugin Notification API**: Connected the Lua `kairo.notify` function to the TUI status bar, allowing plugins to provide visual feedback directly to the user.
* **Fixed Plugin Notifications**: Resolved an issue where plugin notifications were not appearing in the status bar due to missing async message handling.
* **MCP Server Port Control**: Added support for running the built-in MCP server in SSE/HTTP mode on a specific port (`kairo mcp <port>`).
* **MCP Configuration**: Added `mcp_port` setting to `config.toml` and `KAIRO_MCP_PORT` environment variable support for flexible port overrides and auto-start configuration.
* **REAL MCP Server Enhancements**: Transformed the built-in MCP server into a professional-grade implementation with support for Resources (`tasks://all`), Prompts (`manage_tasks`), and expanded Tools (including `kairo_get_task` and `kairo_list_tags`).
* **AI Total App Control**: Updated the AI Assistant's system prompt and tool definitions to enable seamless control over UI themes and Lua plugins.
* **Help Menu Clarity**: Added dedicated keybinding information for the Import/Export menu to the global help screen.

### Fixed

* **Focus Management**: Resolved a bug where typing in the Import/Export file path box would inadvertently trigger global app keybindings.

## [1.3.0]

### Added

* **Import/Export Menu**: Introduced a dedicated menu (accessible via `x`) to easily import and export tasks in multiple formats directly from the TUI.
* **CSV and Text Support**: Added support for `.csv` and `.txt` formats to both the TUI and the CLI API, expanding data portability beyond JSON and Markdown.
* **API-Bound Transitions**: The new menu binds directly to the Kairo API, ensuring consistent data handling and validation between the TUI and headless automation.
* **Dynamic File Path Input**: Users can now specify custom file paths for both imports and exports with real-time feedback and default filename suggestions.
* **Bulk Deletion UI**: Added a quick 'Delete All' action (`a`) to the delete confirmation dialog for rapid workspace clearing.
* **Lua Plugin Themes**: The Lua plugin system now supports curating custom themes. Plugins can return a `themes` table with full control over colors and appearance, which persist across sessions.
* **AI Assistant Panel**: Integrated Gemini (3.1 Flash Lite, 2.5 Flash, 2.0 Flash) (`ctrl+a`) for natural language task management. Create, list, and update tasks using conversational prompts with total tool-calling app control.
* **Live UI Syncing**: AI operations via the assistant panel now trigger live asynchronous UI refreshes (zero restart needed).
* **AI Model Selection**: Users can seamlessly switch between Gemini models live via the Settings TUI (using `left`/`right` arrow keys) or in `config.toml`.
* **Google Search Agent**: Running Kairo with `gemini-2.5-flash-lite` automatically unlocks native Google Search grounding capabilities for web-aware automation.
* **Integrated MCP Server**: Built-in Model Context Protocol server (`kairo mcp`) that exposes your entire task database (including deep metadata like `deadline`, `status`, `tags`, and `priority`) to other AI agents.
* **Settings Reset**: Quickly reset all app configurations back to default inside the settings menu by pressing `r`.
* **API & MCP Theme Control**: Change the entire TUI theme via the headless API (`kairo api set_theme`) or MCP tools (`kairo_set_theme`).
* **API & MCP Plugin Control**: Full management of Lua plugins (list, get, write, delete) through the CLI API and MCP server, enabling AI agents to extend Kairo's functionality.
* **Status Indicators**: The footer now displays a real-time "MCP" pill when the built-in server is active.

### Changed

* **CLI Enhancements**: Updated `kairo import` and `kairo export` commands to support the new `--format [csv|txt]` options.
* **Command Palette**: Added "Import/Export" to the global command palette for quick access.


## [1.2.4]

### Added

* **Vim Navigation**: Added support for `gg` in Vim Mode to instantly jump to the top of the task list, complementing the existing `G` shortcut for bottom navigation.

### Fixed

* **Help Footer Bug**: Resolved a bug where the "Show Help Footer" setting (in both `config.toml` and the settings menu) was being ignored. The footer now correctly hides help keybinding pills when disabled, providing more vertical space for task lists while still retaining critical action prompts (like delete/quit confirmation) for improved usability.

### Changed

* **Keybinding Refinement**: Removed the legacy hardcoded `g` shortcut for plugin reloading in the main list view, as it conflicted with Vim mode navigation. Plugin management remains accessible via the dedicated plugin menu and command palette.

## [1.2.3]

### Added

* **Theme Previews**: The Theme Menu now displays intuitive mini-swatches for every theme, accurately rendering the background, foreground, accent, and success colors side-by-side for flawless visual previewing on any terminal background.
* **Header Breathing Room**: Added a subtle top margin to the header, pushing the "KAIRO" logo and tabs down slightly for a more balanced, uncrowded layout.
* **GitHub Discussions (`u`)**: Added dedicated shortcut to open the project's GitHub Discussions page.
* **Footer UI Update**: Redesigned the footer with individual, circular pill containers (using powerline-style caps) for each keybinding to maximize readability and aesthetic appeal. Keybindings are left-aligned while version and sync status remain anchored to the right.
* **Settings Menu**: Added an interactive settings menu (accessible via `ctrl+s`) to live-configure application settings, with support for live config file watching, reloading, and a shortcut (`g`) to directly open `config.toml` for advanced configuration.
* **Empty State Dashboard**: Transformed the empty home screen into a personal productivity dashboard with compressed, elegant greetings, a minimalist rocket icon, and real-time task completion statistics.
* **Theme Improvements**: Updated the `Nord` theme's muted color to a more prominent tone for improved legibility.

### Fixed

* **Responsive Auto-Resize Engine**: Fully implemented dynamic width constraints across the application. 
    * The Header Tabs now dynamically shrink and truncate titles (`Upc…`) to guarantee they never overflow the window horizontally.
    * The overall Header block and task count pill perfectly anchor to the exact center (`Align(lipgloss.Center)`), surviving aggressive terminal resizing without drifting.
    * The Footer (`render.BarLine`) correctly clips to the terminal width without shattering the layout grid.
* **Tab Switch Panics**: Fixed a crash (`strings: negative Repeat count`) that triggered when switching tabs rapidly during narrow terminal conditions.
* **Menu Box Centering**: Resolved an issue where the Help and Theme menu overlays would drift to the left; they now properly inherit the viewport dimensions and float perfectly dead-center.
* **Help Menu Alignment**: Corrected the text alignment inside the Help box to render cleanly on the left instead of forcing awkward center-justification.
* **Cohesive Pill Caps**: Extended the premium powerline pill styling (`` / ``) to the Header Tabs (including smooth animated bubble transitions), the `DELETE?` / `QUIT?` footer badges, and all Task Priority labels (P0-P3).
* **Linear Rainbow Animation Fix**: Resolved a race condition where toggling the rainbow logo multiple times (or changing other settings) would spawn multiple ticker loops, causing the animation to accelerate. It now maintains a consistent, buttery-smooth frame rate.

### Changed

* **Homebrew Repository Modularity**: Migrated the Homebrew Cask publishing from the primary application repository to a dedicated, independent tap repository (`programmersd21/kairo_tap`) to maintain cleaner git history, modularity, and separation of distribution concerns.

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
