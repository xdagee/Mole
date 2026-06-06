//go:build windows

package platform

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	shell32               = windows.NewLazySystemDLL("shell32.dll")
	procSHEmptyRecycleBin = shell32.NewProc("SHEmptyRecycleBinW")
	procSHQueryRecycleBin = shell32.NewProc("SHQueryRecycleBinW")
)

type SHQUERYRBINFO struct {
	CbSize      uint32
	_           uint32 // Padding for 8-byte alignment on x64
	I64Size     int64
	I64NumItems int64
}

// WindowsPlatform implements the Platform interface for Windows.
type WindowsPlatform struct{}

func newPlatform() Platform {
	return &WindowsPlatform{}
}

// === Path Resolution ===

func (p *WindowsPlatform) UserCacheDir() string {
	if cache := os.Getenv("LOCALAPPDATA"); cache != "" {
		return cache
	}
	return os.Getenv("TEMP")
}

func (p *WindowsPlatform) UserConfigDir() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return appData
	}
	return os.Getenv("LOCALAPPDATA")
}

func (p *WindowsPlatform) UserLogDir() string {
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		return filepath.Join(localAppData, "Logs")
	}
	return os.Getenv("TEMP")
}

func (p *WindowsPlatform) UserTempDir() string {
	if temp := os.Getenv("TEMP"); temp != "" {
		return temp
	}
	return os.Getenv("TMP")
}

func (p *WindowsPlatform) AppDirs() []string {
	progFiles := os.Getenv("ProgramFiles")
	progFilesX86 := os.Getenv("ProgramFiles(x86)")
	localAppData := os.Getenv("LOCALAPPDATA")

	dirs := []string{progFiles}
	if progFilesX86 != "" && progFilesX86 != progFiles {
		dirs = append(dirs, progFilesX86)
	}
	if localAppData != "" {
		dirs = append(dirs, filepath.Join(localAppData, "Programs"))
	}
	return dirs
}

func (p *WindowsPlatform) SystemCacheDirs() []string {
	return []string{
		os.Getenv("TEMP"),
		filepath.Join(os.Getenv("SystemRoot"), "Temp"),
	}
}

func (p *WindowsPlatform) SystemLogDirs() []string {
	return []string{
		filepath.Join(os.Getenv("SystemRoot"), "Logs"),
		filepath.Join(os.Getenv("SystemRoot"), "System32", "winevt", "Logs"),
	}
}

func (p *WindowsPlatform) StartupDirs() []StartupLocation {
	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")
	progData := os.Getenv("ProgramData")

	return []StartupLocation{
		// Startup folders
		{Path: filepath.Join(progData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup"), Type: "directory"},
		{Path: filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup"), Type: "directory"},
		// Local Programs Start Menu
		{Path: filepath.Join(localAppData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup"), Type: "directory"},
		// Registry Run keys
		{RegistryKey: `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, Type: "registry"},
		{RegistryKey: `HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, Type: "registry"},
		{RegistryKey: `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`, Type: "registry"},
		{RegistryKey: `HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`, Type: "registry"},
		// Task Scheduler
		{Path: filepath.Join(os.Getenv("SystemRoot"), "System32", "Tasks"), Type: "scheduler"},
	}
}

func (p *WindowsPlatform) MoleCacheDir() string {
	dir := filepath.Join(p.UserCacheDir(), "mole")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func (p *WindowsPlatform) MoleConfigDir() string {
	dir := filepath.Join(p.UserConfigDir(), "mole")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func (p *WindowsPlatform) MoleLogDir() string {
	dir := filepath.Join(p.UserLogDir(), "mole")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// === App Discovery ===

func (p *WindowsPlatform) ListApps(ctx context.Context) ([]AppInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var apps []AppInfo

	// 1. Scan Start Menu shortcuts
	startMenuPaths := []string{
		filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs"),
	}

	for _, startPath := range startMenuPaths {
		apps = append(apps, p.scanStartMenuDir(ctx, startPath)...)
	}

	// 2. Scan Registry Uninstall keys
	apps = append(apps, p.getRegistryApps()...)

	// 3. Scan Program Files directories
	for _, appDir := range p.AppDirs() {
		apps = append(apps, p.scanAppDir(ctx, appDir)...)
	}

	// Deduplicate by name
	seen := make(map[string]bool)
	var unique []AppInfo
	for _, app := range apps {
		if !seen[app.Name] {
			seen[app.Name] = true
			unique = append(unique, app)
		}
	}

	return unique, nil
}

func (p *WindowsPlatform) scanStartMenuDir(ctx context.Context, dir string) []AppInfo {
	var apps []AppInfo
	entries, err := os.ReadDir(dir)
	if err != nil {
		return apps
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return apps
		default:
		}

		if entry.IsDir() {
			subApps := p.scanStartMenuDir(ctx, filepath.Join(dir, entry.Name()))
			apps = append(apps, subApps...)
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".lnk") &&
			!strings.HasSuffix(strings.ToLower(name), ".exe") {
			continue
		}

		displayName := strings.TrimSuffix(name, filepath.Ext(name))
		apps = append(apps, AppInfo{
			Name:   displayName,
			Path:   filepath.Join(dir, name),
			Size:   -1,
		})
	}

	return apps
}

func (p *WindowsPlatform) scanAppDir(ctx context.Context, dir string) []AppInfo {
	var apps []AppInfo
	entries, err := os.ReadDir(dir)
	if err != nil {
		return apps
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return apps
		default:
		}

		if !entry.IsDir() {
			continue
		}

		appPath := filepath.Join(dir, entry.Name())
		// Check for .exe files in the directory
		appEntries, err := os.ReadDir(appPath)
		if err == nil {
			for _, appEntry := range appEntries {
				if strings.HasSuffix(strings.ToLower(appEntry.Name()), ".exe") {
					apps = append(apps, AppInfo{
						Name:   entry.Name(),
						Path:   appPath,
						Size:   -1,
					})
					break
				}
			}
		}
	}

	return apps
}

func (p *WindowsPlatform) getRegistryApps() []AppInfo {
	var apps []AppInfo

	// Check both HKLM and HKCU uninstall keys
	uninstallPaths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, basePath := range uninstallPaths {
		for _, hive := range []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER} {
			key, err := registry.OpenKey(hive, basePath, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
			if err != nil {
				continue
			}

			subKeys, err := key.ReadSubKeyNames(-1)
			key.Close()
			if err != nil {
				continue
			}

			for _, subKey := range subKeys {
				subKeyPath := basePath + `\` + subKey
				appKey, err := registry.OpenKey(hive, subKeyPath, registry.QUERY_VALUE)
				if err != nil {
					continue
				}

				displayName, _, _ := appKey.GetStringValue("DisplayName")
				if displayName == "" {
					appKey.Close()
					continue
				}

				installLocation, _, _ := appKey.GetStringValue("InstallLocation")
				uninstallString, _, _ := appKey.GetStringValue("UninstallString")
				quietUninstallString, _, _ := appKey.GetStringValue("QuietUninstallString")

				info := AppInfo{
					Name: displayName,
					Size: -1,
				}
				if installLocation != "" {
					info.Path = installLocation
				} else if uninstallString != "" {
					info.Path = uninstallString // Fallback path if none
				}
				
				if quietUninstallString != "" {
					info.UninstallString = quietUninstallString
				} else {
					info.UninstallString = uninstallString
				}

				appKey.Close()
				apps = append(apps, info)
			}
		}
	}

	return apps
}

func (p *WindowsPlatform) GetAppDisplayName(appPath string) string {
	// Try to get display name from executable metadata
	info, err := os.Stat(appPath)
	if err == nil {
		return strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
	}
	return filepath.Base(appPath)
}

func (p *WindowsPlatform) GetAppBundleID(appPath string) string {
	// Windows doesn't have bundle IDs like macOS.
	// Return the base name as a pseudo-identifier.
	return filepath.Base(appPath)
}

func (p *WindowsPlatform) IsAppRunning(appName string) bool {
	// Strip .exe suffix if present
	name := strings.TrimSuffix(appName, ".exe")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use tasklist to check if the process is running, resolving path to system32
	cmd := exec.CommandContext(ctx, getSystem32Path("tasklist.exe"), "/FI", fmt.Sprintf("IMAGENAME eq %s.exe", name), "/NH")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}

	// If the output contains the process name, it's running
	return strings.Contains(strings.ToLower(out.String()), strings.ToLower(name))
}

// === Authentication ===

func (p *WindowsPlatform) IsAdmin() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

func (p *WindowsPlatform) RequireAdmin(ctx context.Context) error {
	if p.IsAdmin() {
		return nil
	}
	// Re-execute with UAC elevation, resolving path to system32 cmd.exe
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	verb := "runas"
	cmd := exec.CommandContext(ctx, getSystem32Path("cmd.exe"), "/C", "start", verb, exe)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// === System Info ===

func (p *WindowsPlatform) OSVersion() OSVersion {
	arch := goArch()
	name := "Windows"
	version := "unknown"
	build := "unknown"

	// Try registry first (most reliable)
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err == nil {
		productName, _, _ := key.GetStringValue("ProductName")
		displayVersion, _, _ := key.GetStringValue("DisplayVersion")
		currentBuild, _, _ := key.GetStringValue("CurrentBuildNumber")
		key.Close()

		if productName != "" {
			name = productName
		}
		if displayVersion != "" {
			version = displayVersion
		}
		if currentBuild != "" {
			build = currentBuild
		}
	}

	// Fallback: use Go's OS version
	if version == "unknown" {
		v := getWindowsVersion()
		if v != "" {
			version = v
		}
	}

	return OSVersion{
		Name:    name,
		Version: version,
		Build:   build,
		Arch:    arch,
	}
}

func getWindowsVersion() string {
	major, minor, build := windows.RtlGetNtVersionNumbers()
	return fmt.Sprintf("%d.%d.%d", major, minor, build)
}

func (p *WindowsPlatform) IsEncrypted() bool {
	// Check BitLocker status, resolving path to system32 manage-bde.exe
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, getSystem32Path("manage-bde.exe"), "-status", os.Getenv("SystemDrive"))
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		output := out.String()
		return strings.Contains(output, "Fully Encrypted") ||
			strings.Contains(output, "Used Space Only Encrypted")
	}
	return false
}

func (p *WindowsPlatform) IsFirewallEnabled() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, getSystem32Path("netsh.exe"), "advfirewall", "show", "currentprofile")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		output := strings.ToLower(out.String())
		return strings.Contains(output, "on") && !strings.Contains(output, "not available")
	}
	return false
}

func (p *WindowsPlatform) IsSIPEnabled() bool {
	// Windows has no direct SIP equivalent.
	// SmartScreen is the closest, but it's not a system integrity feature.
	return false
}

// === System Operations ===

func (p *WindowsPlatform) FlushDNS(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, getSystem32Path("ipconfig.exe"), "/flushdns")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to flush DNS cache: %w", err)
	}
	return nil
}

func (p *WindowsPlatform) EmptyTrash(ctx context.Context) error {
	// Native Win32 API call to silently empty the Recycle Bin on all drives in-process
	ret, _, err := procSHEmptyRecycleBin.Call(0, 0, 0x00000007) // SHERB_NOCONFIRMATION | SHERB_NOPROGRESSUI | SHERB_NOSOUND
	if int32(ret) < 0 {
		return fmt.Errorf("failed to empty recycle bin: HRESULT 0x%x (err: %w)", ret, err)
	}
	return nil
}

func (p *WindowsPlatform) OpenInExplorer(path string, reveal bool) {
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = "C:\\Windows"
	}
	explorer := filepath.Join(systemRoot, "explorer.exe")
	if reveal {
		// Reveal in Explorer
		go exec.Command(explorer, "/select,", path).Run()
	} else {
		// Open normally
		go exec.Command(explorer, path).Run()
	}
}

func (p *WindowsPlatform) PreviewFile(path string) error {
	// Windows has no direct Quick Look equivalent.
	// Open the default viewer instead, resolving path to system32 rundll32.exe
	go exec.Command(getSystem32Path("rundll32.exe"), "shell32.dll,OpenAs_RunDLL", path).Run()
	return nil
}

func (p *WindowsPlatform) GetTrashSize() int64 {
	// Native Win32 API call to query Recycle Bin size across all drives
	var info SHQUERYRBINFO
	info.CbSize = uint32(unsafe.Sizeof(info))
	ret, _, _ := procSHQueryRecycleBin.Call(0, uintptr(unsafe.Pointer(&info)))
	if int32(ret) < 0 {
		return 0
	}
	return info.I64Size
}

// === Helpers ===

func goArch() string {
	arch := os.Getenv("GOARCH")
	if arch != "" {
		return arch
	}
	// Use PROCESSOR_ARCHITECTURE environment variable
	if procArch := os.Getenv("PROCESSOR_ARCHITECTURE"); procArch != "" {
		switch strings.ToLower(procArch) {
		case "amd64":
			return "amd64"
		case "arm64":
			return "arm64"
		case "x86":
			return "386"
		}
		return strings.ToLower(procArch)
	}
	return "unknown"
}

func getSystem32Path(binary string) string {
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = "C:\\Windows"
	}
	return filepath.Join(systemRoot, "System32", binary)
}
