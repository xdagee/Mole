// Package installer implements installer file detection and removal.
// Ported from: bin/installer.sh
package installer

import (
	"context"

	"github.com/tw93/mole/pkg/logutil"
)

// Finder detects and removes installer files.
type Finder struct {
	logger *logutil.Logger
}

// New creates a new Finder.
func New(logger *logutil.Logger) *Finder {
	return &Finder{logger: logger}
}

// Run executes the interactive installer cleanup flow.
func (f *Finder) Run(ctx context.Context, dryRun bool) error {
	// TODO: Phase 3 implementation
	// 1. Scan common download locations for installer files:
	//    - Downloads, Desktop, Documents
	//    - Homebrew cache (Darwin) / Winget cache (Windows)
	//    - iCloud Downloads (Darwin)
	// 2. Detect installer types: .dmg, .pkg, .iso, .xip (Darwin)
	//    .exe, .msi, .iso (Windows)
	// 3. Interactive multi-select with source labels
	// 4. Remove selected installers
	return nil
}
