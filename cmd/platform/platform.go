// Package platform provides OS-specific abstractions for Mole.
// It defines interfaces that are implemented separately for Darwin and Windows,
// allowing the rest of the codebase to remain platform-agnostic.
package platform

import (
	"context"
	"time"
)

// Platform defines the OS-specific operations interface.
// All business logic should call methods on Current rather than
// using OS-specific commands directly.
type Platform interface {
	// === Path Resolution ===

	// UserCacheDir returns the user-level cache directory.
	// Darwin: ~/Library/Caches
	// Windows: %LOCALAPPDATA%\Cache
	UserCacheDir() string

	// UserConfigDir returns the user-level config directory.
	// Darwin: ~/.config
	// Windows: %APPDATA%
	UserConfigDir() string

	// UserLogDir returns the user-level log directory.
	// Darwin: ~/Library/Logs
	// Windows: %LOCALAPPDATA%\Logs
	UserLogDir() string

	// UserTempDir returns the user-level temp directory.
	// Darwin: $TMPDIR
	// Windows: %TEMP%
	UserTempDir() string

	// AppDirs returns directories where applications are installed.
	// Darwin: ["/Applications", "~/Applications"]
	// Windows: ["C:\Program Files", "C:\Program Files (x86)", "%LOCALAPPDATA%\Programs"]
	AppDirs() []string

	// SystemCacheDirs returns system-level cache directories.
	// Darwin: ["/Library/Caches"]
	// Windows: ["%TEMP%", "C:\Windows\Temp"]
	SystemCacheDirs() []string

	// SystemLogDirs returns system-level log directories.
	// Darwin: ["/private/var/log", "/Library/Logs"]
	// Windows: ["%SystemRoot%\Logs"]
	SystemLogDirs() []string

	// StartupDirs returns directories containing startup items.
	// Darwin: ["~/Library/LaunchAgents", "/Library/LaunchAgents", "/Library/LaunchDaemons"]
	// Windows: Registry Run keys, Startup folders, Task Scheduler
	StartupDirs() []StartupLocation

	// MoleCacheDir returns Mole's cache directory.
	// Darwin: ~/.cache/mole
	// Windows: %LOCALAPPDATA%\mole
	MoleCacheDir() string

	// MoleConfigDir returns Mole's config directory.
	// Darwin: ~/.config/mole
	// Windows: %APPDATA%\mole
	MoleConfigDir() string

	// MoleLogDir returns Mole's log directory.
	// Darwin: ~/Library/Logs/mole
	// Windows: %LOCALAPPDATA%\Logs\mole
	MoleLogDir() string

	// === App Discovery ===

	// ListApps returns all installed applications.
	ListApps(ctx context.Context) ([]AppInfo, error)

	// GetAppDisplayName returns the user-facing name for an app.
	GetAppDisplayName(appPath string) string

	// GetAppBundleID returns the bundle identifier for an app.
	// Darwin: reads Info.plist CFBundleIdentifier
	// Windows: reads registry or returns empty
	GetAppBundleID(appPath string) string

	// IsAppRunning checks if an application is currently running.
	IsAppRunning(appName string) bool

	// === Authentication ===

	// IsAdmin returns true if the current process has admin privileges.
	IsAdmin() bool

	// RequireAdmin ensures admin privileges are available.
	// Darwin: checks sudo, may prompt via GUI dialog
	// Windows: checks UAC token, may trigger elevation
	RequireAdmin(ctx context.Context) error

	// === System Info ===

	// OSVersion returns operating system version information.
	OSVersion() OSVersion

	// IsEncrypted returns true if the system drive is encrypted.
	// Darwin: FileVault status
	// Windows: BitLocker status
	IsEncrypted() bool

	// IsFirewallEnabled returns true if the system firewall is enabled.
	IsFirewallEnabled() bool

	// IsSIPEnabled returns true if System Integrity Protection is enabled.
	// Darwin: csrutil status
	// Windows: always returns false (no equivalent)
	IsSIPEnabled() bool

	// === System Operations ===

	// FlushDNS clears the system DNS cache.
	FlushDNS(ctx context.Context) error

	// EmptyTrash empties the system trash/recycle bin.
	EmptyTrash(ctx context.Context) error

	// OpenInExplorer opens a path in the system file explorer.
	// Darwin: open -R (reveal in Finder)
	// Windows: explorer.exe /select
	OpenInExplorer(path string, reveal bool)

	// PreviewFile opens a file in the system quick-look/preview.
	// Darwin: qlmanage -p
	// Windows: not available, returns nil
	PreviewFile(path string) error

	// GetTrashSize returns the size of the trash/recycle bin in bytes.
	GetTrashSize() int64
}

// AppInfo represents an installed application.
type AppInfo struct {
	// Name is the user-facing display name.
	Name string

	// BundleID is the unique identifier (Darwin) or registry key (Windows).
	BundleID string

	// Path is the full filesystem path to the application.
	Path string

	// UninstallString is the command to uninstall the application (Windows).
	UninstallString string

	// Size is the total size in bytes (-1 if unknown).
	Size int64

	// LastUsed is the last time the app was opened (-1 if unknown).
	LastUsed time.Time

	// IsRunning indicates whether the app is currently running.
	IsRunning bool

	// IsBackground is true for background-only apps (daemon-like).
	IsBackground bool
}

// OSVersion contains operating system version information.
type OSVersion struct {
	// Name is the human-readable OS name (e.g., "macOS Sonoma", "Windows 11").
	Name string

	// Version is the version string (e.g., "14.5", "10.0.22631").
	Version string

	// Build is the build number.
	Build string

	// Arch is the CPU architecture ("arm64", "amd64").
	Arch string
}

// StartupLocation represents a location where startup items are stored.
type StartupLocation struct {
	// Path is the filesystem path (for directory-based locations).
	Path string

	// RegistryKey is the registry path (for Windows registry-based locations).
	RegistryKey string

	// Type describes the location type ("directory", "registry", "scheduler").
	Type string
}

// Current is the active platform implementation, resolved at init time.
var Current Platform

func init() {
	Current = newPlatform()
}
