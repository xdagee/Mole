// Package optimize implements system optimization and maintenance tasks.
// Ported from: bin/optimize.sh, lib/optimize/*.sh
package optimize

import (
	"context"
	"os/exec"

	"github.com/tw93/mole/pkg/logutil"
	"github.com/tw93/mole/cmd/platform"
)

// Optimizer handles system optimization tasks.
type Optimizer struct {
	logger *logutil.Logger
}

// New creates a new Optimizer.
func New(logger *logutil.Logger) *Optimizer {
	return &Optimizer{logger: logger}
}

// Run executes all optimization tasks.
func (o *Optimizer) Run(ctx context.Context, dryRun bool) error {
	o.logger.Info("Starting optimization tasks...")

	if dryRun {
		o.logger.Info("Dry run: Would flush DNS, run DISM cleanup, and run SFC scan")
		return nil
	}

	// 1. Flush DNS
	o.logger.Info("Flushing DNS cache...")
	if err := platform.Current.FlushDNS(ctx); err != nil {
		o.logger.Warning("DNS flush failed: %v", err)
	} else {
		o.logger.Success("DNS cache flushed")
	}

	// 2. DISM Cleanup
	o.logger.Info("Running DISM component cleanup (this may take a while)...")
	dismCmd := exec.CommandContext(ctx, "dism.exe", "/online", "/Cleanup-Image", "/StartComponentCleanup")
	if err := dismCmd.Run(); err != nil {
		o.logger.Warning("DISM cleanup failed: %v", err)
	} else {
		o.logger.Success("DISM cleanup completed")
	}

	// 3. SFC Scan
	o.logger.Info("Running System File Checker (SFC) scan...")
	sfcCmd := exec.CommandContext(ctx, "sfc.exe", "/scannow")
	if err := sfcCmd.Run(); err != nil {
		o.logger.Warning("SFC scan failed or found unfixable errors: %v", err)
	} else {
		o.logger.Success("SFC scan completed")
	}

	return nil
}
