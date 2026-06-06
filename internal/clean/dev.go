package clean

import (
	"context"
	"os"
	"path/filepath"

	"github.com/tw93/mole/pkg/fsutil"
	"github.com/tw93/mole/pkg/logutil"
)

// DevCacheModule cleans developer tool caches.
type DevCacheModule struct {
	logger *logutil.Logger
}

func (m *DevCacheModule) Name() string { return "Developer caches" }

func (m *DevCacheModule) getPaths() []string {
	userProfile := os.Getenv("USERPROFILE")
	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")

	paths := []string{}
	if appData != "" {
		paths = append(paths, filepath.Join(appData, "npm-cache"))
	}
	if localAppData != "" {
		paths = append(paths, filepath.Join(localAppData, "pip", "Cache"))
		paths = append(paths, filepath.Join(localAppData, "Yarn", "Cache"))
		paths = append(paths, filepath.Join(localAppData, "go-build"))
	}
	if userProfile != "" {
		paths = append(paths, filepath.Join(userProfile, ".gradle", "caches"))
		paths = append(paths, filepath.Join(userProfile, ".cargo", "registry", "cache"))
		// Some tools use standard unix paths even on Windows
		paths = append(paths, filepath.Join(userProfile, ".npm", "_cacache"))
	}
	return paths
}

func (m *DevCacheModule) Run(ctx context.Context, dryRun bool) (int, int64, error) {
	var totalBytes int64
	var totalFiles int

	for _, path := range m.getPaths() {
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
					if m.logger != nil {
						m.logger.Ops("DELETE", path, size)
					}
				}
			}
		}
	}
	return totalFiles, totalBytes, nil
}

func (m *DevCacheModule) DryRunPreview(ctx context.Context) ([]PreviewEntry, error) {
	var entries []PreviewEntry
	for _, path := range m.getPaths() {
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			size, _ := fsutil.GetDirectorySize(ctx, path)
			if size > 0 {
				entries = append(entries, PreviewEntry{
					Path: path,
					Size: size,
					Type: "dev-cache",
				})
			}
		}
	}
	return entries, nil
}
