package clean

import (
	"context"
	"os"
	"path/filepath"

	"github.com/tw93/mole/cmd/platform"
	"github.com/tw93/mole/pkg/logutil"
)

// SystemCacheModule cleans system-wide temporary files.
type SystemCacheModule struct {
	logger *logutil.Logger
}

func (m *SystemCacheModule) Name() string { return "System caches" }

func (m *SystemCacheModule) getPaths() []string {
	var paths []string
	p := platform.Current
	if p != nil {
		paths = append(paths, p.SystemCacheDirs()...)
	}
	
	sysDrive := os.Getenv("SystemDrive")
	if sysDrive == "" {
		sysDrive = "C:"
	}
	// Windows Update cache
	paths = append(paths, filepath.Join(sysDrive, "\\Windows", "SoftwareDistribution", "Download"))
	
	return paths
}

func (m *SystemCacheModule) Run(ctx context.Context, dryRun bool) (int, int64, error) {
	return cleanPaths(ctx, m.getPaths(), "system-cache", m.logger, dryRun)
}

func (m *SystemCacheModule) DryRunPreview(ctx context.Context) ([]PreviewEntry, error) {
	return previewPaths(ctx, m.getPaths(), "system-cache")
}
