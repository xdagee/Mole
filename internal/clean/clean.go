// Package clean implements system cache and data cleanup.
// Ported from: bin/clean.sh, lib/clean/*.sh
package clean

import (
	"context"

	"github.com/tw93/mole/cmd/platform"
	"github.com/tw93/mole/pkg/config"
	"github.com/tw93/mole/pkg/logutil"
	"github.com/tw93/mole/pkg/ui"
)

// Module defines a cleanup module that can be run independently.
type Module interface {
	// Name returns the human-readable name of this module.
	Name() string

	// Run executes the cleanup and returns files cleaned and bytes freed.
	Run(ctx context.Context, dryRun bool) (filesCleaned int, bytesFreed int64, err error)

	// DryRunPreview returns what would be cleaned without making changes.
	DryRunPreview(ctx context.Context) (entries []PreviewEntry, err error)
}

// PreviewEntry represents a file/directory that would be cleaned.
type PreviewEntry struct {
	Path string
	Size int64
	Type string // "cache", "log", "temp", etc.
}

// Run executes all cleanup modules.
func Run(ctx context.Context, cfg *config.Config, logger *logutil.Logger, dryRun bool) error {
	modules := []Module{
		&UserCacheModule{logger: logger},
		&BrowserCacheModule{logger: logger},
		&DevCacheModule{logger: logger},
		&SystemCacheModule{logger: logger},
		&AppCacheModule{logger: logger},
		&TrashModule{logger: logger},
	}

	totalFiles := 0
	totalBytes := int64(0)

	for _, module := range modules {
		if cfg.IsWhitelisted(module.Name()) {
			logger.Info("Skipping whitelisted module: %s", module.Name())
			continue
		}

		if dryRun {
			entries, err := module.DryRunPreview(ctx)
			if err != nil {
				logger.Warning("Module %s preview failed: %v", module.Name(), err)
				continue
			}
			for _, entry := range entries {
				logger.Info("  %s %s (%s)", ui.IconArrow, entry.Path, ui.FormatBytes(entry.Size))
			}
			continue
		}

		files, bytes, err := module.Run(ctx, dryRun)
		if err != nil {
			logger.Error("Module %s failed: %v", module.Name(), err)
			continue
		}

		totalFiles += files
		totalBytes += bytes
		logger.Success("Module %s: %d files, %s freed", module.Name(), files, ui.FormatBytes(bytes))
	}

	logger.Info("Total: %d files cleaned, %s freed", totalFiles, ui.FormatBytes(totalBytes))
	return nil
}

// AppCacheModule cleans application specific caches (placeholder for Milestone 4/5 integration).
type AppCacheModule struct{ logger *logutil.Logger }

func (m *AppCacheModule) Name() string                          { return "App caches" }
func (m *AppCacheModule) Run(ctx context.Context, dryRun bool) (int, int64, error) {
	return 0, 0, nil
}
func (m *AppCacheModule) DryRunPreview(ctx context.Context) ([]PreviewEntry, error) {
	return nil, nil
}

// TrashModule cleans the recycle bin.
type TrashModule struct{ logger *logutil.Logger }

func (m *TrashModule) Name() string                          { return "Trash" }
func (m *TrashModule) Run(ctx context.Context, dryRun bool) (int, int64, error) {
	size := platform.Current.GetTrashSize()
	if size > 0 && !dryRun {
		err := platform.Current.EmptyTrash(ctx)
		if err == nil {
			return 1, size, nil
		}
		return 0, 0, err
	}
	return 0, 0, nil
}
func (m *TrashModule) DryRunPreview(ctx context.Context) ([]PreviewEntry, error) {
	size := platform.Current.GetTrashSize()
	if size > 0 {
		return []PreviewEntry{{Path: "Recycle Bin", Size: size, Type: "trash"}}, nil
	}
	return nil, nil
}
