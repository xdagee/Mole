package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/tw93/mole/cmd/platform"
	"github.com/tw93/mole/internal/analyze"
	"github.com/tw93/mole/internal/clean"
	"github.com/tw93/mole/internal/optimize"
	"github.com/tw93/mole/internal/purge"
	"github.com/tw93/mole/internal/status"
	"github.com/tw93/mole/internal/uninstall"
	"github.com/tw93/mole/pkg/config"
	"github.com/tw93/mole/pkg/logutil"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// EventWriter implements io.Writer to forward logs to Wails runtime.
type EventWriter struct {
	ctx context.Context
}

func (e *EventWriter) Write(p []byte) (n int, err error) {
	if e.ctx != nil {
		runtime.EventsEmit(e.ctx, "log", string(p))
	}
	return len(p), nil
}

// App struct
type App struct {
	ctx    context.Context
	logger *logutil.Logger

	mu           sync.Mutex
	activeCancel context.CancelFunc

	statusCollector *status.Collector
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger = logutil.NewStreamLogger(&EventWriter{ctx: ctx})
	a.statusCollector = status.NewCollector(status.ProcessWatchOptions{})
}

// startOperation sets up a cancellable context
func (a *App) startOperation() (context.Context, func()) {
	a.mu.Lock()
	defer a.mu.Unlock()

	ctx, cancel := context.WithCancel(a.ctx)
	a.activeCancel = cancel

	return ctx, func() {
		a.mu.Lock()
		a.activeCancel = nil
		a.mu.Unlock()
		cancel()
	}
}

// CancelOperation aborts the currently running operation.
func (a *App) CancelOperation() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.activeCancel != nil {
		a.activeCancel()
		a.activeCancel = nil
		a.logger.Warning("Operation cancelled by user.")
	}
}

// GetStatus returns the latest system metrics snapshot.
func (a *App) GetStatus() (status.MetricsSnapshot, error) {
	snap, err := a.statusCollector.Collect()
	if err != nil {
		if a.logger != nil {
			a.logger.Warning("GetStatus partial failure: %v", err)
		}
	}
	// Always return nil error so Wails resolves the Promise with the partial snapshot
	return snap, nil
}

// AnalyzeDisk scans a directory and returns its usage tree.
func (a *App) AnalyzeDisk(target string) (*analyze.Node, error) {
	ctx, done := a.startOperation()
	defer done()

	if target == "" {
		target = "C:\\" // Default to C: on Windows if empty
	}
	return analyze.ScanDirectory(ctx, target)
}

// RunClean executes system cleanup
func (a *App) RunClean(dryRun bool) string {
	ctx, done := a.startOperation()
	defer done()

	cfg := config.New()
	err := clean.Run(ctx, cfg, a.logger, dryRun)
	if err != nil {
		if err == context.Canceled {
			return "Cleanup cancelled"
		}
		return fmt.Sprintf("Error: %v", err)
	}
	return "Cleanup completed successfully"
}

// PreviewClean returns what would be cleaned without executing it.
func (a *App) PreviewClean() []clean.PreviewEntry {
	ctx, done := a.startOperation()
	defer done()

	cfg := config.New()
	
	// Temporarily create modules directly (as Run handles this internally)
	modules := []clean.Module{
		&clean.UserCacheModule{},
		&clean.BrowserCacheModule{},
		&clean.DevCacheModule{},
		&clean.SystemCacheModule{},
		&clean.AppCacheModule{},
		&clean.TrashModule{},
	}

	var allEntries []clean.PreviewEntry
	for _, mod := range modules {
		if cfg.IsWhitelisted(mod.Name()) {
			continue
		}
		entries, err := mod.DryRunPreview(ctx)
		if err == nil {
			allEntries = append(allEntries, entries...)
		}
	}
	return allEntries
}

// ListApps returns the list of installed applications
func (a *App) ListApps() []platform.AppInfo {
	ctx, done := a.startOperation()
	defer done()

	apps, err := platform.Current.ListApps(ctx)
	if err != nil {
		a.logger.Error("Error listing apps: %v", err)
		return []platform.AppInfo{}
	}
	return apps
}

// RunUninstall uninstalls a specific application
func (a *App) RunUninstall(appName string, dryRun bool) string {
	ctx, done := a.startOperation()
	defer done()

	apps, err := platform.Current.ListApps(ctx)
	if err != nil {
		return fmt.Sprintf("Error listing apps: %v", err)
	}

	var targetApp *platform.AppInfo
	for _, app := range apps {
		if strings.EqualFold(app.Name, appName) {
			targetApp = &app
			break
		}
	}

	if targetApp == nil {
		return "Error: Application not found"
	}

	uninstaller := uninstall.New(a.logger)
	err = uninstaller.Execute(ctx, *targetApp, dryRun)
	if err != nil {
		if err == context.Canceled {
			return "Uninstall cancelled"
		}
		return fmt.Sprintf("Error uninstalling: %v", err)
	}
	return "Uninstall completed successfully"
}

// RunOptimize executes system optimization
func (a *App) RunOptimize(dryRun bool) string {
	ctx, done := a.startOperation()
	defer done()

	optimizer := optimize.New(a.logger)
	err := optimizer.Run(ctx, dryRun)
	if err != nil {
		if err == context.Canceled {
			return "Optimization cancelled"
		}
		return fmt.Sprintf("Error: %v", err)
	}
	return "Optimization completed successfully"
}

// RunPurge cleans project artifacts
func (a *App) RunPurge(dryRun bool) string {
	ctx, done := a.startOperation()
	defer done()

	purger := purge.New(a.logger)
	err := purger.Run(ctx, dryRun)
	if err != nil {
		if err == context.Canceled {
			return "Purge cancelled"
		}
		return fmt.Sprintf("Error: %v", err)
	}
	return "Purge completed successfully"
}
