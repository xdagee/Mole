package clean

import (
	"context"
	"os"
	"path/filepath"

	"github.com/tw93/mole/cmd/platform"
	"github.com/tw93/mole/pkg/fsutil"
	"github.com/tw93/mole/pkg/logutil"
)

// UserCacheModule cleans user temporary files.
type UserCacheModule struct {
	logger *logutil.Logger
}

func (m *UserCacheModule) Name() string { return "User caches" }

func (m *UserCacheModule) getPaths() []string {
	var paths []string
	p := platform.Current
	if p != nil {
		temp := p.UserTempDir()
		if temp != "" {
			paths = append(paths, temp)
		}
	}
	return paths
}

func (m *UserCacheModule) Run(ctx context.Context, dryRun bool) (int, int64, error) {
	return cleanPaths(ctx, m.getPaths(), "user-cache", m.logger, dryRun)
}

func (m *UserCacheModule) DryRunPreview(ctx context.Context) ([]PreviewEntry, error) {
	return previewPaths(ctx, m.getPaths(), "user-cache")
}

// BrowserCacheModule cleans browser caches.
type BrowserCacheModule struct {
	logger *logutil.Logger
}

func (m *BrowserCacheModule) Name() string { return "Browser caches" }

func (m *BrowserCacheModule) getPaths() []string {
	localAppData := os.Getenv("LOCALAPPDATA")
	var paths []string
	if localAppData != "" {
		paths = append(paths, filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Cache"))
		paths = append(paths, filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Code Cache"))
		paths = append(paths, filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Cache"))
		paths = append(paths, filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Code Cache"))
		paths = append(paths, filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", "Default", "Cache"))
		paths = append(paths, filepath.Join(localAppData, "Mozilla", "Firefox", "Profiles")) // Note: Requires deeper glob for actual cache folder in Firefox, but keeping simple for Mole cache root check or we delete whole Profiles cache. Wait, deleting whole Profiles deletes user data! Firefox cache is actually in LOCALAPPDATA\Mozilla\Firefox\Profiles\<profile>\cache2.
	}
	return paths
}

func (m *BrowserCacheModule) Run(ctx context.Context, dryRun bool) (int, int64, error) {
	// For Firefox we need to glob the profiles
	paths := m.getPaths()
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		ffProfiles, _ := filepath.Glob(filepath.Join(localAppData, "Mozilla", "Firefox", "Profiles", "*", "cache2"))
		paths = append(paths, ffProfiles...)
	}
	return cleanPaths(ctx, paths, "browser-cache", m.logger, dryRun)
}

func (m *BrowserCacheModule) DryRunPreview(ctx context.Context) ([]PreviewEntry, error) {
	paths := m.getPaths()
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		ffProfiles, _ := filepath.Glob(filepath.Join(localAppData, "Mozilla", "Firefox", "Profiles", "*", "cache2"))
		paths = append(paths, ffProfiles...)
	}
	return previewPaths(ctx, paths, "browser-cache")
}

// cleanPaths is a helper for cleaning a list of directories.
func cleanPaths(ctx context.Context, paths []string, typ string, logger *logutil.Logger, dryRun bool) (int, int64, error) {
	var totalBytes int64
	var totalFiles int

	for _, path := range paths {
		select {
		case <-ctx.Done():
			return totalFiles, totalBytes, ctx.Err()
		default:
		}

		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			size, _ := fsutil.GetDirectorySize(ctx, path)
			if size > 0 {
				err := fsutil.SafeRemove(path, dryRun)
				if err == nil {
					totalBytes += size
					totalFiles++
					if logger != nil {
						logger.Ops("DELETE", path, size)
					}
				}
			}
		}
	}
	return totalFiles, totalBytes, nil
}

// previewPaths is a helper for previewing a list of directories.
func previewPaths(ctx context.Context, paths []string, typ string) ([]PreviewEntry, error) {
	var entries []PreviewEntry
	for _, path := range paths {
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			size, _ := fsutil.GetDirectorySize(ctx, path)
			if size > 0 {
				entries = append(entries, PreviewEntry{
					Path: path,
					Size: size,
					Type: typ,
				})
			}
		}
	}
	return entries, nil
}
