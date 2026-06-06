// Package purge implements project build artifact cleanup.
// Ported from: bin/purge.sh, lib/clean/project.sh
package purge

import (
	"context"

	"github.com/tw93/mole/pkg/logutil"
)

// Purger handles project artifact removal.
type Purger struct {
	logger *logutil.Logger
}

// New creates a new Purger.
func New(logger *logutil.Logger) *Purger {
	return &Purger{logger: logger}
}

// Run executes the interactive purge flow.
func (p *Purger) Run(ctx context.Context, dryRun bool) error {
	p.logger.Info("Starting purge scan...")
	paths := DefaultPaths()
	if len(paths) == 0 {
		p.logger.Info("No default project paths found.")
		return nil
	}
	
	results, err := Scan(ctx, paths)
	if err != nil {
		return err
	}
	
	if len(results) == 0 {
		p.logger.Info("No purgeable artifacts found.")
		return nil
	}
	
	for targetName, res := range results {
		p.logger.Info("Target: %s, Found: %d items, Total Size: %d bytes", targetName, len(res.Items), res.TotalSize)
	}
	
	// TODO: Wire actual BubbleTea UI for multi-selection.
	return nil
}
