// Package manage implements management utilities: whitelist, update, completion.
// Ported from: lib/manage/*.sh, bin/completion.sh, bin/update.sh, bin/remove.sh
package manage

import (
	"github.com/tw93/mole/pkg/logutil"
)

// Manager handles whitelist, update, and completion management.
type Manager struct {
	logger *logutil.Logger
}

// New creates a new Manager.
func New(logger *logutil.Logger) *Manager {
	return &Manager{logger: logger}
}

// ManageWhitelist handles whitelist editing.
func (m *Manager) ManageWhitelist() error {
	// TODO: Open whitelist file in $EDITOR
	return nil
}

// ManagePurgePaths handles custom purge path configuration.
func (m *Manager) ManagePurgePaths() error {
	// TODO: Open purge_paths file in $EDITOR
	return nil
}

// Update checks for and applies Mole updates.
func (m *Manager) Update() error {
	// TODO: Check GitHub releases, download and replace binary
	return nil
}

// Remove uninstalls Mole from the system.
func (m *Manager) Remove() error {
	// TODO: Remove binary, config files, cache, logs
	return nil
}

// Completion generates and installs shell completion scripts.
func (m *Manager) Completion(shell string, dryRun bool) error {
	// TODO: Generate completion for bash, zsh, fish, PowerShell
	return nil
}
