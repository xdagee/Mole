package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProtectedPaths returns a list of paths that Mole should never delete.
func ProtectedPaths() []string {
	sysDrive := os.Getenv("SystemDrive")
	if sysDrive == "" {
		sysDrive = "C:"
	}
	appData := os.Getenv("APPDATA")
	
	paths := []string{
		filepath.Join(sysDrive, "\\"),
		filepath.Join(sysDrive, "\\Windows"),
		filepath.Join(sysDrive, "\\Program Files"),
		filepath.Join(sysDrive, "\\Program Files (x86)"),
		filepath.Join(sysDrive, "\\ProgramData", "Microsoft"),
		filepath.Join(sysDrive, "\\Users", "Default"),
	}

	if appData != "" {
		paths = append(paths, filepath.Join(appData, "Microsoft", "Windows"))
	}

	return paths
}

// IsProtectedPath checks if a path is protected.
func IsProtectedPath(path string) bool {
	cleanPath := filepath.Clean(path)
	cleanPathLower := strings.ToLower(cleanPath)

	for _, protected := range ProtectedPaths() {
		protectedClean := strings.ToLower(filepath.Clean(protected))
		if cleanPathLower == protectedClean {
			return true
		}
		// Protect parent directories of system roots (e.g. don't delete C:\ if trying to delete C:\Windows)
		// Protect subdirectories of strictly protected paths if explicitly defined, 
		// though typically safe_remove only blocks the exact root or checks prefix.
		// For Mole, we block exact matches and prefix matches for highly sensitive roots.
		if strings.HasPrefix(cleanPathLower, protectedClean+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// SafeRemove attempts to remove a file or directory safely.
// dryRun controls if it actually deletes or just returns success.
func SafeRemove(path string, dryRun bool) error {
	if IsProtectedPath(path) {
		return fmt.Errorf("refusing to delete protected path: %s", path)
	}

	if dryRun {
		return nil
	}

	// Use os.RemoveAll to handle directories and files.
	// Windows will naturally throw an error if the file is locked 
	// (ERROR_SHARING_VIOLATION), which covers our lsof checks.
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove %s: %w", path, err)
	}
	return nil
}
