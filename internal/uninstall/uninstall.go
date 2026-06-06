// Package uninstall implements smart application uninstallation.
// Ported from: bin/uninstall.sh, lib/uninstall/*.sh
package uninstall

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tw93/mole/cmd/platform"
	"github.com/tw93/mole/pkg/logutil"
)

// Uninstaller handles application removal and related file cleanup.
type Uninstaller struct {
	logger *logutil.Logger
}

// New creates a new Uninstaller.
func New(logger *logutil.Logger) *Uninstaller {
	return &Uninstaller{logger: logger}
}

// Execute runs the uninstallation command for a specific app.
func (u *Uninstaller) Execute(ctx context.Context, app platform.AppInfo, dryRun bool) error {
	u.logger.Info("Attempting to uninstall: %s", app.Name)

	if app.UninstallString == "" {
		return fmt.Errorf("no uninstall string found for %s", app.Name)
	}

	u.logger.Info("Uninstall command: %s", app.UninstallString)

	if dryRun {
		return nil
	}

	parts, err := resolveUninstallCommand(app.UninstallString)
	if err != nil || len(parts) == 0 {
		return fmt.Errorf("invalid uninstall string: %w", err)
	}

	// Wait for the uninstaller if it's msiexec. Some uninstallers spawn another process.
	// For now, simple execution is fine.
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("uninstall failed: %v, output: %s", err, string(out))
	}

	u.logger.Success("Successfully uninstalled %s", app.Name)
	return nil
}

// resolveUninstallCommand parses a Windows command line string securely.
// It handles quoted paths and searches the filesystem for unquoted paths with spaces.
func resolveUninstallCommand(cmdLine string) ([]string, error) {
	cmdLine = strings.TrimSpace(cmdLine)
	if len(cmdLine) == 0 {
		return nil, fmt.Errorf("empty command line")
	}

	// 1. Quoted executable path
	if cmdLine[0] == '"' {
		end := strings.Index(cmdLine[1:], "\"")
		if end != -1 {
			execPath := cmdLine[1 : end+1]
			argsStr := cmdLine[end+2:]
			return append([]string{execPath}, strings.Fields(argsStr)...), nil
		}
	}

	// 2. Unquoted path with spaces. Search left-to-right (accumulating parts)
	// or right-to-left checking for an existing file.
	// e.g. C:\Program Files\My App\uninstall.exe /S
	parts := strings.Split(cmdLine, " ")
	for i := len(parts); i > 0; i-- {
		candidatePath := strings.Join(parts[:i], " ")
		candidatePath = strings.Trim(candidatePath, "\"'")
		if stat, err := os.Stat(candidatePath); err == nil && !stat.IsDir() {
			var args []string
			for _, arg := range parts[i:] {
				if trimArg := strings.TrimSpace(arg); trimArg != "" {
					args = append(args, trimArg)
				}
			}
			return append([]string{candidatePath}, args...), nil
		}
	}

	// Fallback to simple fields split
	return strings.Fields(cmdLine), nil
}
