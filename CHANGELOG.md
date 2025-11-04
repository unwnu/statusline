# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-11-04

### Added
- Initial release of statusline tool for Claude Code
- Git status line display with color-coded branch icon
  - Green: clean repository
  - Yellow: tracked changes
  - Red: untracked files
- Upstream tracking with ahead/behind indicators (↑/↓)
- Fetch interval caching with git reflog for optimized performance
- Bold branch icon for better visibility
- Cross-platform support (Linux, macOS, Windows)
- Environment variables for configuration:
  - `STATUSLINE_NO_COLOR` - disable colors
  - `STATUSLINE_FETCH` - enable upstream fetching
  - `STATUSLINE_FETCH_INTERVAL` - configure fetch interval (default: 30 minutes)

### Fixed
- Use --dry-run for fetch to prevent blocking git operations
- Add --no-progress flag to prevent auth prompts during fetch
- README.md naming for goreleaser compatibility

### Documentation
- Cross-platform examples for Claude Code integration
- Installation and configuration instructions
- Environment variable documentation

[1.0.0]: https://github.com/unwnu/statusline/releases/tag/v1.0.0
