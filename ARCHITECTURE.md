# Mole — All-Go Cross-Platform Architecture Design

## 1. Goal

Rewrite Mole as a **single Go binary** that works natively on **macOS and Windows x64**, replacing all Bash/PowerShell shell scripts with Go code while keeping the existing Go TUI components (`cmd/analyze`, `cmd/status`).

---

## 2. Current State

| Layer | macOS (main) | Windows (upstream/windows) |
|-------|-------------|---------------------------|
| **Entry point** | `mole` (Bash) | `mole.ps1` (PowerShell) |
| **Commands** | `bin/*.sh` (10 scripts) | `bin/*.ps1` (8 scripts) |
| **Libraries** | `lib/*.sh` (35 scripts) | `lib/*.ps1` (portered) |
| **TUI** | Go (`cmd/analyze/`, `cmd/status/`) | Go (ported with build tags) |
| **Tests** | BATS (39 `.bats` files) | Pester (PowerShell `.Tests.ps1`) |
| **Packaging** | Homebrew | Chocolatey, Scoop, Winget |

### Problem
Two separate codebases to maintain. Shell scripts are inherently platform-bound.

---

## 3. Target Architecture

```
Mole/
├── cmd/
│   ├── mole/                    # Single unified binary (main.go)
│   ├── analyze/                 # TUI disk analyzer (existing, cross-platform)
│   ├── status/                  # TUI system monitor (existing, cross-platform)
│   └── platform/                # Platform abstraction layer (NEW)
│       ├── platform.go          # Interfaces
│       ├── paths.go             # OS-specific path resolution
│       ├── commands.go          # OS-specific command execution
│       ├── auth.go              # sudo (macOS) / UAC (Windows)
│       ├── app_discovery.go     # Spotlight (macOS) / Registry+Start Menu (Win)
│       ├── system_info.go       # OS version, architecture detection
│       ├── platform_darwin.go   # macOS implementations
│       └── platform_windows.go  # Windows implementations
├── internal/                    # Core business logic (NEW)
│   ├── clean/                   # Cleaning modules
│   │   ├── clean.go             # Shared interfaces
│   │   ├── caches.go            # Cache cleanup
│   │   ├── apps.go              # Orphaned app data
│   │   ├── dev.go               # Developer tool caches
│   │   ├── system.go            # System-level cleanup
│   │   └── user.go              # User data cleanup
│   ├── uninstall/               # App uninstaller
│   │   ├── uninstall.go         # Shared interfaces
│   │   ├── scanner.go           # App discovery
│   │   ├── remover.go           # File removal logic
│   │   └── services.go          # LaunchAgents / Services / Registry
│   ├── optimize/                # System optimization
│   │   ├── optimize.go          # Shared interfaces
│   │   ├── tasks.go             # Optimization tasks
│   │   └── maintenance.go       # Preference repair
│   ├── purge/                   # Project artifact purge
│   │   ├── purge.go             # Shared interfaces
│   │   ├── scanner.go           # Project discovery
│   │   └── targets.go           # Known artifact patterns
│   ├── installer/               # Installer file cleanup
│   │   └── installer.go         # Installer detection and removal
│   ├── check/                   # System health checks
│   │   ├── check.go             # Shared interfaces
│   │   ├── security.go          # Firewall, encryption, SIP
│   │   ├── updates.go           # OS and app updates
│   │   └── dev.go               # Dev environment checks
│   └── manage/                  # Management utilities
│       ├── whitelist.go         # Whitelist management
│       ├── update.go            # Self-update
│       └── completion.go        # Shell tab completion
├── pkg/                         # Shared utilities
│   ├── ui/                      # Terminal UI (Bubble Tea + custom)
│   │   ├── menu.go              # Multi-select menus
│   │   ├── spinner.go           # Progress spinners
│   │   └── table.go             # Formatted tables
│   ├── fsutil/                  # File system utilities
│   │   ├── safe_remove.go       # Protected path validation
│   │   ├── size.go              # Size calculation
│   │   └── glob.go              # Pattern matching
│   ├── logutil/                 # Logging with rotation
│   │   └── logger.go
│   └── config/                  # Configuration management
│       └── config.go
├── tests/                       # Go tests
├── packaging/                   # Release packaging
│   ├── homebrew/                # Homebrew formula
│   ├── chocolatey/              # Chocolatey package
│   ├── scoop/                   # Scoop manifest
│   └── winget/                  # Winget manifest
├── Makefile                     # Build system
├── go.mod
└── QWEN.md
```

---

## 4. Platform Abstraction Layer

### 4.1 Core Interfaces (`cmd/platform/platform.go`)

```go
package platform

import "context"

// Platform defines the OS-specific operations interface.
type Platform interface {
    // Paths
    UserCacheDir() string
    UserConfigDir() string
    UserLogDir() string
    UserTempDir() string
    AppDirs() []string           // /Applications or C:\Program Files
    SystemCacheDirs() []string
    SystemLogDirs() []string
    StartupDirs() []string       // LaunchAgents or Registry Run

    // Authentication
    IsAdmin() bool
    RequireAdmin(ctx context.Context) error  // sudo or UAC

    // App Discovery
    ListApps() ([]AppInfo, error)
    GetAppDisplayName(bundleID string) string
    GetAppBundleID(appPath string) string
    IsAppRunning(bundleID string) bool
    FindAppsByName(name string) []AppInfo

    // System Info
    OSVersion() OSVersion
    IsAdminAvailable() bool
    IsEncrypted() bool           // FileVault / BitLocker
    IsFirewallEnabled() bool
    IsSIPEnabled() bool          // macOS only, returns false on Windows

    // System Operations
    FlushDNS(ctx context.Context) error
    RebuildLaunchServices(ctx context.Context) error  // macOS only
    FlushFontCache(ctx context.Context) error
    RepairPermissions(ctx context.Context) error
    EmptyTrash(ctx context.Context) error
    OpenInExplorer(path string)
}

type AppInfo struct {
    Name       string
    BundleID   string
    Path       string
    Size       int64
    LastUsed   time.Time
    IsRunning  bool
}

type OSVersion struct {
    Name    string  // "macOS Sonoma" / "Windows 11"
    Version string
    Build   string
    Arch    string  // "arm64" / "amd64"
}
```

### 4.2 Platform Resolution

```go
// platform.go — resolved at init time
var Current Platform

func init() {
    switch runtime.GOOS {
    case "darwin":
        Current = &DarwinPlatform{}
    case "windows":
        Current = &WindowsPlatform{}
    default:
        Current = &UnsupportedPlatform{}
    }
}
```

### 4.3 Path Mapping

| Concept | macOS | Windows |
|---------|-------|---------|
| User cache | `~/Library/Caches` | `%LOCALAPPDATA%\Cache` |
| User config | `~/.config` | `%APPDATA%` |
| User logs | `~/Library/Logs` | `%LOCALAPPDATA%\Logs` |
| User temp | `$TMPDIR` | `%TEMP%` |
| App dirs | `/Applications`, `~/Applications` | `C:\Program Files`, `C:\Program Files (x86)`, `%LOCALAPPDATA%\Programs` |
| System cache | `/Library/Caches` | `%TEMP%`, `C:\Windows\Temp` |
| System logs | `/private/var/log`, `/Library/Logs` | `%SystemRoot%\Logs` |
| Startup items | `~/Library/LaunchAgents`, `/Library/LaunchDaemons` | Registry `Run` keys, `Startup` folder, Task Scheduler |
| Config storage | `~/.config/mole/` | `%APPDATA%\mole\` |

### 4.4 Command Equivalents

| Operation | macOS | Windows |
|-----------|-------|---------|
| Admin check | `sudo -n true` | `whoami /all` → S-1-5-32-544 |
| Admin elevation | `sudo` | UAC (`Start-Process -Verb RunAs` via `os/exec`) |
| App list | Spotlight `mdfind`, `/Applications` scan | Start Menu `.lnk` scan, Registry `Uninstall` keys |
| Bundle ID | `Info.plist` → `CFBundleIdentifier` | Registry `ProductCode`, exe metadata |
| Running check | `pgrep`, `osascript`, `lsappinfo` | `Get-Process`, WMI `Win32_Process` |
| DNS flush | `dscacheutil -flushcache`, `mDNSResponder` | `ipconfig /flushdns` |
| Trash empty | AppleScript → Finder | `Shell.Application` → `NameSpace(10).Items()` |
| Disk encryption | `fdesetup status` | `manage-bde -status` |
| Firewall | `socketfilterfw --getglobalstate` | `netsh advfirewall show currentprofile` |
| SIP | `csrutil status` | N/A (SmartScreen equivalent) |
| File size | `stat -f%z` (BSD) | `os.Stat()` (Go stdlib handles both) |
| Plist editing | `PlistBuddy`, `plutil` | Registry API, JSON/YAML configs |
| Service management | `launchctl load/unload` | `sc.exe`, `Get-Service`, Task Scheduler |

---

## 5. Module Migration Map

### 5.1 `bin/*.sh` → Go Commands

| Shell Script | Go Package | Complexity | Notes |
|-------------|-----------|-----------|-------|
| `bin/clean.sh` | `internal/clean/` | **Very High** | Orchestrates 8 sub-modules. Biggest migration effort. |
| `bin/uninstall.sh` | `internal/uninstall/` | **Very High** | Spotlight metadata → Windows Registry/Start Menu. |
| `bin/optimize.sh` | `internal/optimize/` | **High** | 20+ tasks, many macOS-specific (DNS, font cache, LaunchServices). |
| `bin/purge.sh` | `internal/purge/` | **Medium** | Mostly path scanning — already cross-platform. |
| `bin/installer.sh` | `internal/installer/` | **Medium** | Path scanning + file type detection. |
| `bin/check.sh` | `internal/check/` | **High** | Security checks (Firewall, FileVault, SIP) need Windows equivalents. |
| `bin/touchid.sh` | `cmd/platform/auth.go` | **High (macOS-only)** | Windows: UAC + Windows Hello integration. |
| `bin/completion.sh` | `internal/manage/completion.go` | **Low** | Shell detection logic is portable. |
| `bin/analyze.sh` | `cmd/analyze/` | **Done** | Already Go, uses build tags. |
| `bin/status.sh` | `cmd/status/` | **Done** | Already Go, uses build tags. |

### 5.2 `lib/*.sh` → Go Packages

| Shell Library | Go Package | Complexity | Notes |
|--------------|-----------|-----------|-------|
| `lib/core/base.sh` | `cmd/platform/` | **High** | Constants, colors, icons, architecture detection. |
| `lib/core/file_ops.sh` | `pkg/fsutil/` | **Medium** | Safe remove, protected paths, size calculation. |
| `lib/core/log.sh` | `pkg/logutil/` | **Low** | Logging with rotation — pure Go. |
| `lib/core/sudo.sh` | `cmd/platform/auth.go` | **High** | Touch ID ↔ Windows Hello, GUI password prompts. |
| `lib/core/timeout.sh` | Go `context.WithTimeout` | **Done** | Go stdlib handles this natively. |
| `lib/core/ui.sh` | `pkg/ui/` | **Medium** | Terminal control — Bubble Tea handles most of it. |
| `lib/core/common.sh` | `internal/` shared | **Low** | Dock management (macOS-only), Homebrew update. |
| `lib/core/app_protection.sh` | `pkg/fsutil/` | **Medium** | Protected app lists — data-only, portable. |
| `lib/core/help.sh` | `cmd/mole/` | **Low** | Help text — pure strings. |
| `lib/clean/*.sh` (8 files) | `internal/clean/` | **Very High** | Hundreds of macOS-specific paths. |
| `lib/uninstall/*.sh` (2 files) | `internal/uninstall/` | **High** | LaunchServices, Login Items, Homebrew casks. |
| `lib/optimize/*.sh` (2 files) | `internal/optimize/` | **High** | DNS, font cache, SQLite vacuum, periodic scripts. |
| `lib/check/*.sh` (3 files) | `internal/check/` | **High** | Security, updates, dev environment. |
| `lib/manage/*.sh` (4 files) | `internal/manage/` | **Medium** | Whitelist, update, purge paths. |
| `lib/ui/*.sh` (3 files) | `pkg/ui/` | **Medium** | Menus — migrate to Bubble Tea components. |

---

## 6. Implementation Phases

### Phase 1: Platform Abstraction Layer (Weeks 1-2)
**Goal**: All business logic calls interfaces, not OS-specific code.

1. Define `Platform` interface in `cmd/platform/platform.go`
2. Implement `DarwinPlatform` (copy logic from existing shell scripts)
3. Implement `WindowsPlatform` (new, use Windows APIs and WMI)
4. Write tests: verify each platform returns correct paths
5. Move all constants (colors, icons, protected paths) into platform-specific files

**Deliverable**: `cmd/platform/` with full interface coverage, both platforms compiling.

### Phase 2: Core Utilities (Weeks 2-3)
**Goal**: Port `lib/core/` to Go packages.

1. `pkg/fsutil/` — safe file operations, protected paths, size calculation
2. `pkg/logutil/` — logging with rotation, operation audit
3. `pkg/ui/` — terminal UI (menus, spinners) using Bubble Tea
4. `pkg/config/` — configuration management (JSON/TOML files)
5. Replace `lib/core/timeout.sh` with Go `context.WithTimeout` (trivial)

**Deliverable**: All core utilities in Go, tested, no shell dependencies.

### Phase 3: Command Modules — Easiest First (Weeks 3-5)
**Goal**: Port commands with least platform dependency first.

| Order | Command | Reason |
|-------|---------|--------|
| 1 | `purge` | Pure file scanning, minimal platform dependency |
| 2 | `installer` | File type scanning, portable logic |
| 3 | `check` | Mix of portable (disk, memory) and platform-specific (security) |
| 4 | `manage` | Whitelist, update, completion — mostly config editing |
| 5 | `clean` | Biggest module — 8 sub-modules, hundreds of paths |
| 6 | `uninstall` | Complex app discovery, LaunchServices, services |
| 7 | `optimize` | Most platform-specific operations |

**Deliverable**: All commands working on both macOS and Windows.

### Phase 4: TUI Enhancement (Week 5-6)
**Goal**: Ensure `cmd/analyze/` and `cmd/status/` work perfectly on Windows.

1. Update build tags for proper platform separation
2. Add Windows-specific metrics (page file, WMI data)
3. Test terminal rendering on Windows Terminal, CMD, PowerShell
4. Add Windows-specific health indicators

### Phase 5: Build & CI/CD (Week 6-7)
**Goal**: Automated cross-platform builds and releases.

1. **Makefile**:
   - `make build` → current platform
   - `make release-darwin-amd64` → macOS Intel
   - `make release-darwin-arm64` → macOS Apple Silicon
   - `make release-windows-amd64` → Windows x64

2. **GitHub Actions**:
   - `check.yml` — lint, vet, format (both platforms)
   - `test.yml` — Go tests (both platforms)
   - `release.yml` — multi-platform release with:
     - macOS: Homebrew tap
     - Windows: Chocolatey, Scoop, Winget

3. **Packaging**:
   - Port existing `packaging/chocolatey/`, `packaging/scoop/`, `packaging/winget/` from Windows branch
   - Update Homebrew formula for new single-binary structure

**Deliverable**: `go install github.com/tw93/mole@latest` works on both platforms.

### Phase 6: Testing & Polish (Week 7-8)
1. Port BATS tests → Go tests
2. Manual testing on Windows 11 (Terminal, CMD, PowerShell)
3. Manual testing on macOS (Intel + Apple Silicon)
4. Security audit (protected paths, admin elevation)
5. Performance benchmarking

---

## 7. Makefile Design

```makefile
# Unified Makefile for Mole (macOS + Windows)

.PHONY: all build clean test release-darwin-amd64 release-darwin-arm64 release-windows-amd64

GO ?= go
BIN_DIR := bin
LDFLAGS := -s -w

all: build

build:
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/mole$(EXE) ./cmd/mole/
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/analyze$(EXE) ./cmd/analyze/
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/status$(EXE) ./cmd/status/

test:
	$(GO) test -v ./cmd/... ./internal/... ./pkg/...

clean:
	rm -f $(BIN_DIR)/mole* $(BIN_DIR)/analyze* $(BIN_DIR)/status*

# Release builds (run on native architectures for CGO support)
release-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/mole-darwin-amd64 ./cmd/mole/

release-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/mole-darwin-arm64 ./cmd/mole/

release-windows-amd64:
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/mole-windows-amd64.exe ./cmd/mole/

release: release-darwin-amd64 release-darwin-arm64 release-windows-amd64
```

---

## 8. Windows-Specific Implementation Notes

### 8.1 App Discovery
```go
// Windows: Scan Start Menu + Registry Uninstall keys
func (p *WindowsPlatform) ListApps() ([]AppInfo, error) {
    // 1. Scan Start Menu shortcuts
    startMenuPaths := []string{
        filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs"),
        filepath.Join("C:", "ProgramData", "Microsoft", "Windows", "Start Menu", "Programs"),
    }
    // 2. Scan Registry: HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall
    // 3. Scan HKCU equivalent
    // 4. Scan common install locations
}
```

### 8.2 Service Management
```go
// Windows: Task Scheduler + Registry Run keys instead of LaunchAgents
func (p *WindowsPlatform) ListStartupItems() ([]StartupItem, error) {
    // 1. Registry Run keys (HKLM + HKCU)
    // 2. Startup folders
    // 3. Task Scheduler tasks
    // 4. Windows Services
}
```

### 8.3 Admin Elevation
```go
// Windows: UAC elevation via re-exec with verb "runas"
func (p *WindowsPlatform) RequireAdmin(ctx context.Context) error {
    if p.IsAdmin() {
        return nil
    }
    // Re-exec current process with UAC prompt
    verb := "runas"
    exe, _ := os.Executable()
    cmd := exec.CommandContext(ctx, "cmd", "/C", "start", verb, exe)
    return cmd.Run()
}
```

### 8.4 Windows Cleanup Targets
| Category | Windows Equivalent |
|----------|-------------------|
| `~/Library/Caches` | `%LOCALAPPDATA%\Temp`, `%TEMP%` |
| `~/Library/Logs` | `%LOCALAPPDATA%\Logs` |
| Homebrew cache | Winget cache (`%LOCALAPPDATA%\Packages\Microsoft.DesktopAppInstaller_*`), Scoop cache |
| Xcode DerivedData | Visual Studio components, `.vs/`, `bin/`, `obj/` |
| npm/pnpm/bun caches | Same paths (`%USERPROFILE%\.npm`, etc.) — already cross-platform |
| Time Machine snapshots | Volume Shadow Copy, File History |
| LaunchAgents | Registry Run, Task Scheduler, Services |
| Spotlight index | Windows Search index |

---

## 9. What Stays, What Changes, What's New

### Stays the Same
- `cmd/analyze/` — Already cross-platform via build tags. Minor Windows path adjustments.
- `cmd/status/` — gopsutil already supports Windows. Add Windows-specific metrics.
- CLI interface (`mo clean`, `mo uninstall`, etc.) — Same commands, different backends.
- Safety-first defaults — Protected paths, dry-run, confirmation prompts.
- Operation logging — Same concept, Windows paths.

### Changes
- Shell scripts → Go code (all of `bin/`, `lib/`)
- Bash TUI menus → Bubble Tea components
- `~/Library/*` paths → Platform-resolved paths
- `sudo` → `sudo` (macOS) / UAC (Windows)
- Homebrew → Homebrew (macOS) / Winget+Scoop (Windows)

### New
- Windows-specific cleanup targets (Temp files, Winget cache, Windows Update cleanup)
- Windows-specific security checks (BitLocker, SmartScreen, UAC level)
- Windows-specific optimizations (DISM cleanup, SFC scan, defrag analysis)
- Chocolatey/Scoop/Winget packaging

---

## 10. Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Windows path handling bugs | High | Extensive test coverage on path resolution, protected paths |
| Admin elevation failures | High | Graceful fallback to manual instructions |
| False-positive file deletion | Critical | Protected path validation, dry-run mandatory, allowlists |
| Performance regression on large scans | Medium | Use Go goroutines, file caching, parallel scanning |
| Terminal rendering on Windows | Medium | Test on Windows Terminal, CMD, PowerShell separately |
| Breaking existing macOS users | High | Keep CLI interface identical, run parallel BATS tests |

---

## 11. Estimated Scope

| Phase | Effort | Key Deliverable |
|-------|--------|----------------|
| Phase 1: Platform Abstraction | 2 weeks | `cmd/platform/` with interfaces |
| Phase 2: Core Utilities | 1-2 weeks | `pkg/` packages |
| Phase 3: Command Modules | 3 weeks | All commands in Go |
| Phase 4: TUI Enhancement | 1 week | Windows-ready TUI |
| Phase 5: Build & CI/CD | 1 week | Cross-platform releases |
| Phase 6: Testing & Polish | 1-2 weeks | Full test coverage |
| **Total** | **9-11 weeks** | Single binary, macOS + Windows |

---

## 12. Current Implementation Status

### Phase 1: Platform Abstraction Layer ✅ COMPLETE
- [x] `cmd/platform/platform.go` — `Platform` interface with 25+ methods
- [x] `cmd/platform/platform_darwin.go` — macOS implementation (build tag: `//go:build darwin`)
- [x] `cmd/platform/platform_windows.go` — Windows implementation (build tag: `//go:build windows`)
- Verified: `mo uninstall` finds 557+ apps on Windows via Registry + Start Menu scanning
- Verified: `mo check` reports platform info, firewall, encryption status correctly

### Phase 2: Core Utilities ✅ COMPLETE
- [x] `pkg/fsutil/safe_remove.go` — Protected path validation, safe remove, find-and-delete
- [x] `pkg/logutil/logger.go` — Multi-target logging with rotation (main, ops, debug)
- [x] `pkg/config/config.go` — Config management (whitelist, purge paths, glob matching)
- [x] `pkg/ui/components.go` — Bubble Tea menu, spinner, progress bar, styling

### Phase 3: Command Modules ✅ COMPLETE
- [x] `cmd/mole/main.go` — Unified entry point, all commands routed, flag parsing
- [x] `internal/clean/` — Core clean engine with user, dev, system cache sweepers
- [x] `internal/uninstall/` — Active registry uninstaller executing silent/quiet removals
- [x] `internal/optimize/` — DNS flush, DISM cleanup, and SFC scans
- [x] `internal/purge/` — Target scanner for node_modules, target, venv, gradle, etc.
- [x] `internal/installer/` — Executable and package installer file detection
- [x] `internal/check/` — Security checks, memory pressure, and updates
- [x] `internal/manage/` — Whitelist and custom path configuration

### Makefile ✅ UPDATED
- [x] Cross-platform build targets (darwin-amd64, darwin-arm64, windows-amd64)
- [x] `make build` for local, `make release` for all platforms

### Build Verification
```
$ go vet ./cmd/... ./pkg/... ./internal/...   # PASS
$ go test ./...                                # PASS
$ go build ./...                               # PASS
$ bin/mole.exe --help                          # Shows full help
$ bin/mole.exe check                           # Runs security and memory checks
$ bin/mole.exe uninstall                       # Lists installed applications
```

### Next Steps
1. **GitHub Actions Integration** — Fully verify the release pipeline from tag triggers (`release.yml`).
2. **User Acceptance Testing** — Deploy the production-ready binary to client setups.