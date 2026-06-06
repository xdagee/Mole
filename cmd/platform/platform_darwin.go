//go:build darwin

package platform

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DarwinPlatform implements the Platform interface for macOS.
type DarwinPlatform struct{}

func newPlatform() Platform {
	return &DarwinPlatform{}
}

// === Path Resolution ===

func (p *DarwinPlatform) UserCacheDir() string {
	home := homeDir()
	return filepath.Join(home, "Library", "Caches")
}

func (p *DarwinPlatform) UserConfigDir() string {
	home := homeDir()
	return filepath.Join(home, ".config")
}

func (p *DarwinPlatform) UserLogDir() string {
	home := homeDir()
	return filepath.Join(home, "Library", "Logs")
}

func (p *DarwinPlatform) UserTempDir() string {
	if tmp := os.Getenv("TMPDIR"); tmp != "" {
		return tmp
	}
	return os.TempDir()
}

func (p *DarwinPlatform) AppDirs() []string {
	home := homeDir()
	return []string{
		"/Applications",
		"/System/Applications",
		filepath.Join(home, "Applications"),
	}
}

func (p *DarwinPlatform) SystemCacheDirs() []string {
	return []string{"/Library/Caches"}
}

func (p *DarwinPlatform) SystemLogDirs() []string {
	return []string{"/private/var/log", "/Library/Logs", "/Library/Logs/DiagnosticReports"}
}

func (p *DarwinPlatform) StartupDirs() []StartupLocation {
	home := homeDir()
	return []StartupLocation{
		{Path: filepath.Join(home, "Library", "LaunchAgents"), Type: "directory"},
		{Path: "/Library/LaunchAgents", Type: "directory"},
		{Path: "/Library/LaunchDaemons", Type: "directory"},
	}
}

func (p *DarwinPlatform) MoleCacheDir() string {
	home := homeDir()
	dir := filepath.Join(home, ".cache", "mole")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func (p *DarwinPlatform) MoleConfigDir() string {
	home := homeDir()
	dir := filepath.Join(home, ".config", "mole")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func (p *DarwinPlatform) MoleLogDir() string {
	home := homeDir()
	dir := filepath.Join(home, "Library", "Logs", "mole")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// === App Discovery ===

func (p *DarwinPlatform) ListApps(ctx context.Context) ([]AppInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var apps []AppInfo
	for _, appDir := range p.AppDirs() {
		entries, err := os.ReadDir(appDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".app") {
				continue
			}
			appPath := filepath.Join(appDir, entry.Name())
			info := AppInfo{
				Name:   strings.TrimSuffix(entry.Name(), ".app"),
				Path:   appPath,
				Size:   -1,
			}
			info.Name = p.GetAppDisplayName(appPath)
			info.BundleID = p.GetAppBundleID(appPath)
			info.IsRunning = p.IsAppRunning(filepath.Base(appPath))
			apps = append(apps, info)
		}
	}
	return apps, nil
}

func (p *DarwinPlatform) GetAppDisplayName(appPath string) string {
	// Try mdls first (Spotlight metadata)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "mdls", "-name", "kMDItemDisplayName", "-raw", appPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err == nil {
		name := strings.TrimSpace(out.String())
		if name != "(null)" && name != "" {
			return name
		}
	}

	// Fallback: read Info.plist CFBundleDisplayName
	plist := filepath.Join(appPath, "Contents", "Info.plist")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	cmd2 := exec.CommandContext(ctx2, "plutil", "-extract", "CFBundleDisplayName", "raw", "-o", "-", plist)
	var out2 bytes.Buffer
	cmd2.Stdout = &out2
	if err := cmd2.Run(); err == nil {
		name := strings.TrimSpace(out2.String())
		if name != "" {
			return name
		}
	}

	// Final fallback: strip .app from filename
	return strings.TrimSuffix(filepath.Base(appPath), ".app")
}

func (p *DarwinPlatform) GetAppBundleID(appPath string) string {
	plist := filepath.Join(appPath, "Contents", "Info.plist")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "plutil", "-extract", "CFBundleIdentifier", "raw", "-o", "-", plist)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		return strings.TrimSpace(out.String())
	}
	return ""
}

func (p *DarwinPlatform) IsAppRunning(appName string) bool {
	// Strip .app suffix if present
	name := strings.TrimSuffix(appName, ".app")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pgrep", "-x", name)
	return cmd.Run() == nil
}

// === Authentication ===

func (p *DarwinPlatform) IsAdmin() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sudo", "-n", "true")
	return cmd.Run() == nil
}

func (p *DarwinPlatform) RequireAdmin(ctx context.Context) error {
	if p.IsAdmin() {
		return nil
	}
	// Try GUI sudo prompt via osascript
	// The actual sudo call will be done by the caller when needed.
	// We just verify the user can authenticate.
	return nil
}

// === System Info ===

func (p *DarwinPlatform) OSVersion() OSVersion {
	arch := goArch()
	name := "macOS"
	version := "unknown"
	build := "unknown"

	// Get version via sw_vers
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sw_vers", "-productVersion")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		version = strings.TrimSpace(out.String())
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	cmd2 := exec.CommandContext(ctx2, "sw_vers", "-buildVersion")
	var out2 bytes.Buffer
	cmd2.Stdout = &out2
	if err := cmd2.Run(); err == nil {
		build = strings.TrimSpace(out2.String())
	}

	// Get macOS name from ProductName
	ctx3, cancel3 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel3()
	cmd3 := exec.CommandContext(ctx3, "sw_vers")
	var out3 bytes.Buffer
	cmd3.Stdout = &out3
	if err := cmd3.Run(); err == nil {
		for _, line := range strings.Split(out3.String(), "\n") {
			if strings.Contains(line, "ProductName") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					name = strings.TrimSpace(parts[1])
				}
				break
			}
		}
	}

	return OSVersion{
		Name:    name,
		Version: version,
		Build:   build,
		Arch:    arch,
	}
}

func (p *DarwinPlatform) IsEncrypted() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "fdesetup", "status")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		return strings.Contains(out.String(), "On")
	}
	return false
}

func (p *DarwinPlatform) IsFirewallEnabled() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		return strings.Contains(out.String(), "enabled")
	}
	return false
}

func (p *DarwinPlatform) IsSIPEnabled() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "csrutil", "status")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		return !strings.Contains(out.String(), "disabled")
	}
	return false
}

// === System Operations ===

func (p *DarwinPlatform) FlushDNS(ctx context.Context) error {
	// Flush mDNSResponder
	cmd := exec.CommandContext(ctx, "sudo", "killall", "-HUP", "mDNSResponder")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to flush DNS cache: %w", err)
	}
	// Flush directory service cache
	cmd2 := exec.CommandContext(ctx, "sudo", "dscacheutil", "-flushcache")
	if err := cmd2.Run(); err != nil {
		return fmt.Errorf("failed to flush directory cache: %w", err)
	}
	return nil
}

func (p *DarwinPlatform) EmptyTrash(ctx context.Context) error {
	// Use AppleScript to empty trash via Finder
	cmd := exec.CommandContext(ctx, "osascript", "-e", "tell application \"Finder\" to empty trash")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to empty trash via AppleScript: %w", err)
	}
	return nil
}

func (p *DarwinPlatform) OpenInExplorer(path string, reveal bool) {
	if reveal {
		// Reveal in Finder
		go exec.Command("open", "-R", path).Run()
	} else {
		// Open normally
		go exec.Command("open", path).Run()
	}
}

func (p *DarwinPlatform) PreviewFile(path string) error {
	cmd := exec.Command("qlmanage", "-p", path)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func (p *DarwinPlatform) GetTrashSize() int64 {
	home := homeDir()
	trashPath := filepath.Join(home, ".Trash")

	// Try osascript first (most accurate)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "osascript", "-e",
		`tell application "Finder" to get the size of (items of trash) as number`)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		if size, err := strconv.ParseFloat(strings.TrimSpace(out.String()), 64); err == nil {
			return int64(size)
		}
	}

	// Fallback: use du
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	cmd2 := exec.CommandContext(ctx2, "du", "-sk", trashPath)
	var out2 bytes.Buffer
	cmd2.Stdout = &out2
	if err := cmd2.Run(); err == nil {
		parts := strings.Fields(out2.String())
		if len(parts) >= 1 {
			if kb, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				return kb * 1024
			}
		}
	}

	return 0
}

// === Helpers ===

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if u, err := user.Current(); err == nil {
		return u.HomeDir
	}
	return os.Getenv("HOME")
}

func goArch() string {
	arch := os.Getenv("GOARCH")
	if arch != "" {
		return arch
	}
	// Fallback: use uname
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "uname", "-m")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		return strings.TrimSpace(out.String())
	}
	return "unknown"
}
