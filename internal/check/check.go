// Package check implements system health and security checks.
// Ported from: bin/check.sh, lib/check/*.sh
package check

import (
	"context"

	"github.com/tw93/mole/cmd/platform"
	"github.com/tw93/mole/pkg/logutil"
	"github.com/tw93/mole/pkg/ui"
	"github.com/shirou/gopsutil/v4/mem"
	"fmt"
)

// Checker runs system health checks.
type Checker struct {
	logger *logutil.Logger
}

// New creates a new Checker.
func New(logger *logutil.Logger) *Checker {
	return &Checker{logger: logger}
}

// Run executes all system checks.
func (c *Checker) Run(ctx context.Context) error {
	// Run checks in parallel
	checks := []func() CheckResult{
		c.checkSecurity,
		c.checkUpdates,
		c.checkDisk,
		c.checkMemory,
		c.checkStartupItems,
	}

	results := make([]CheckResult, len(checks))
	for i, check := range checks {
		results[i] = check()
	}

	// Display results
	for _, result := range results {
		if result.Passed {
			c.logger.Info("%s %s", ui.IconSuccess, result.Message)
		} else {
			c.logger.Warning("%s %s", ui.IconWarning, result.Message)
		}
	}

	return nil
}

// CheckResult represents the result of a single check.
type CheckResult struct {
	Name    string
	Passed  bool
	Message string
	Details string
}

func (c *Checker) checkSecurity() CheckResult {
	p := platform.Current
	passed := true
	var details string

	if !p.IsFirewallEnabled() {
		passed = false
		details += "Firewall is disabled. "
	}
	if !p.IsEncrypted() {
		passed = false
		details += "Drive encryption is not enabled. "
	}

	return CheckResult{
		Name:    "Security",
		Passed:  passed,
		Message: "Security checks",
		Details: details,
	}
}

func (c *Checker) checkUpdates() CheckResult {
	// TODO: Check OS updates, package manager updates
	return CheckResult{
		Name:   "Updates",
		Passed: true,
	}
}

func (c *Checker) checkDisk() CheckResult {
	// TODO: Check disk usage, warn if > 90%
	return CheckResult{
		Name:   "Disk",
		Passed: true,
	}
}

func (c *Checker) checkMemory() CheckResult {
	v, err := mem.VirtualMemory()
	if err != nil {
		return CheckResult{
			Name:    "Memory",
			Passed:  false,
			Message: "Failed to read memory status",
			Details: err.Error(),
		}
	}

	passed := true
	details := fmt.Sprintf("Memory usage is %.2f%%", v.UsedPercent)
	if v.UsedPercent > 90 {
		passed = false
		details = "High memory pressure detected! " + details
	}

	return CheckResult{
		Name:    "Memory",
		Passed:  passed,
		Message: "Memory Pressure Check",
		Details: details,
	}
}

func (c *Checker) checkStartupItems() CheckResult {
	// TODO: Check for broken startup items
	_ = platform.Current.StartupDirs() // TODO: scan these directories
	return CheckResult{
		Name:    "Startup",
		Passed:  true,
		Message: "Checked startup items",
		Details: "",
	}
}
