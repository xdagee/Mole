# Mole Agent Guide

This file is the shared source of truth for any AI agent working on this repo (Claude Code, Codex, etc.). `CLAUDE.md` is a symlink to this file. Put machine-specific or personal overrides in `AGENTS.local.md` / `CLAUDE.local.md`; both are gitignored.

## Project

Mole is a cross-platform system cleanup and optimization tool written in Go, supporting macOS and Windows x64. It replaces legacy shell-based components with a compiled, single-binary architecture. It performs file cleanup, app protection checks, and maintenance tasks, prioritizing safety and validation over speed.

## Repository Map

- `cmd/mole/` - Unified CLI entrypoint (`main.go`). Handles CLI flag parsing, help text generation, interactive Bubble Tea menu routing, and module execution.
- `cmd/platform/` - OS abstraction layer (`platform.go`, `platform_darwin.go`, `platform_windows.go`). Handles path resolution, admin elevation checks, Registry scans, WMI queries, DNS flushes, and Recycle Bin/Trash emptying.
- `cmd/analyze/` - Go disk-analysis TUI (Bubble Tea). Holds the update chain and file scanning/deletion views.
- `cmd/gui/` - Native Wails-based graphical user interface (React/TypeScript). Features interactive Disk Analyzer and System Status dashboards.
- `internal/status/` - System metrics collector. Gathers cross-platform CPU, RAM, disk, power, network, and process metrics.
- `internal/clean/` - Cleanup engine. Includes modular sweepers:
  - `user.go` - Windows Temp, browser caches (Chrome, Edge, Brave, Firefox).
  - `dev.go` - Developer tool caches (npm, pip, gradle, rust/cargo).
  - `system.go` - System caches (SoftwareDistribution, Windows Update).
  - `clean.go` - Module orchestration interface.
- `internal/uninstall/` - App discovery and uninstaller (`uninstall.go`). Queries Uninstall Registry keys, executes uninstall commands (e.g. `QuietUninstallString`), and cleans leftover directories.
- `internal/optimize/` - System maintenance tasks (`optimize.go`). Flushes DNS, runs DISM Component Cleanup, and triggers System File Checker (SFC) scans.
- `internal/purge/` - Project build artifact scanner (`purge.go`, `targets.go`). Finds and purges `node_modules`, `target`, `build`, `.gradle`, etc.
- `internal/installer/` - Installer file discovery and cleanup (`installer.go`). Detects `.exe`, `.msi`, `.dmg`, `.pkg`, `.iso` files.
- `internal/check/` - Health and security checks (`check.go`). Validates BitLocker, Firewall, Memory pressure, and Startup items.
- `pkg/fsutil/` - Safe filesystem operations:
  - `safe_remove.go` - Safe path validation against system critical directories (preventing recursive wipes).
  - `glob.go` - High-performance concurrent directory walker that explicitly skips NTFS junctions/symlinks to prevent cyclic recursion loops.
- `pkg/logutil/` - Logging wrapper (`logger.go`) writing rotation-aware console, ops, and debug logs.
- `pkg/config/` - Whitelist and settings configuration.
- `pkg/ui/` - Reusable CLI UI elements, spinners, and Bubble Tea components.
- `scripts/` - CI check, test, build, and release helpers.
- `docs/SECURITY_DESIGN.md` - Security design specifications for path validation, app protection, and reparse points.

## Commands

```bash
# Run Go unit tests
go test ./...

# Build binary for current platform
make build

# Clean build artifacts
make clean

# Run linters and checks
make verify

# Dry run mode via CLI flag
./bin/mole.exe clean --dry-run
./bin/mole.exe purge --dry-run
```

Public docs and examples should prefer the installed `mo` command. Use `./bin/mole` (or `./bin/mole.exe`) when verifying source-tree behavior before installation.

## Critical Safety Rules

- Never use raw `os.RemoveAll` or `os.Remove` in cleanup modules; always route deletions through `fsutil.SafeRemove()` from `pkg/fsutil/safe_remove.go`.
- Never modify protected paths like system directories (`C:\Windows`, `/System`, etc.).
- Skip NTFS junction points and symbolic links during directory walks to prevent infinite recursion loops on Windows.
- Route user-selected cleanup through system Trash/Recycle Bin where the project expects recoverability, especially for TUI-driven ad hoc cleanup.
- Never let verification block on UAC prompt elevation, AppleScript, or macOS authorization prompts unless the task explicitly targets authentication.
- Check `fsutil.IsProtectedPath()` before executing deletion flows.
- Use `MOLE_DRY_RUN=1` or `--dry-run` to verify behavior before committing destructive changes.
- Preserve rotation-aware operation logging (`OPS` logs) to audit all deletions.

## Working Rules

- Keep AI-tool cache cleanup conservative. Claude Code, opencode, Copilot CLI, Zed, Warp, Ghostty, and similar developer tools may have active versions, configs, or credentials that must not be removed accidentally.
- Prefer targeted Go unit tests (e.g. `go test -v ./pkg/fsutil/...`) during development.
- Do not add AI attribution trailers to commits.

## Hotspot Ownership

These directories and files hold core logic:
- `cmd/platform/platform_windows.go` owns Windows-specific registry scans, UAC elevation, WMI checks, and paths.
- `pkg/fsutil/safe_remove.go` owns path validation and system protection.
- `pkg/fsutil/glob.go` owns concurrent walking and NTFS junction filters.
- `internal/clean/user.go` and `internal/clean/dev.go` own cache sweepers.
- `cmd/mole/main.go` handles CLI orchestration and Bubble Tea command routing.
- `cmd/gui/app.go` owns the Wails desktop application backend bridging.
- `internal/status/metrics.go` owns system-wide telemetry collection.

## Verification

- Go changes: run `go test ./...` and `go vet ./...`.
- Safety verification: run with `--dry-run` or `--debug` to confirm paths match expected categories without deleting files.

## Release

Tag-driven flow. The `release.yml` workflow watches `'V*'` tag pushes (capital `V`), builds amd64 and arm64 binaries for macOS and Windows, generates `SHA256SUMS`, attaches build provenance, and creates the GitHub Release.

### Pre-flight checklist

1. `grep '^VERSION=' cmd/mole/main.go` or version constant matches the new version.
2. `SECURITY.md` or release logs reflect the correct version and date.
3. `git status -s` is clean.
4. `go test ./...` and `make verify` both pass.
5. All binaries compile successfully across target platforms (`make release-amd64`).
