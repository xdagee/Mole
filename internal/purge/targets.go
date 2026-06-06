package purge

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tw93/mole/pkg/fsutil"
)

// Target represents a specific type of project artifact.
type Target struct {
	Name        string
	Directories []string
}

// KnownTargets defines the common project build artifacts we scan for.
var KnownTargets = []Target{
	{Name: "Node.js", Directories: []string{"node_modules", ".npm"}},
	{Name: "Rust", Directories: []string{"target"}},
	{Name: "Go", Directories: []string{"bin"}},
	{Name: "Java/Maven/Gradle", Directories: []string{"target", "build", ".gradle"}},
	{Name: "Python", Directories: []string{"__pycache__", ".pytest_cache", ".venv"}},
	{Name: "Xcode/Swift", Directories: []string{".build", "DerivedData"}},
	{Name: "Generic Build", Directories: []string{"dist", "out"}},
}

// DefaultPaths returns the default search paths for projects on Windows.
func DefaultPaths() []string {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return nil
	}
	
	candidates := []string{
		filepath.Join(userProfile, "Projects"),
		filepath.Join(userProfile, "projects"),
		filepath.Join(userProfile, "source", "repos"),
		filepath.Join(userProfile, "Documents", "GitHub"),
		filepath.Join(userProfile, "dev"),
	}

	var valid []string
	for _, p := range candidates {
		if stat, err := os.Stat(p); err == nil && stat.IsDir() {
			valid = append(valid, p)
		}
	}
	return valid
}

// ScanResults holds the results of scanning for project artifacts.
type ScanResults struct {
	Items []fsutil.ScanResult
	TotalSize int64
}

// Scan finds all matching directories.
func Scan(ctx context.Context, paths []string) (map[string]ScanResults, error) {
	results := make(map[string]ScanResults)
	
	// Collect all target directories
	var allDirs []string
	dirToTarget := make(map[string]string)
	
	for _, t := range KnownTargets {
		for _, d := range t.Directories {
			allDirs = append(allDirs, d)
			dirToTarget[d] = t.Name
		}
	}

	for _, d := range allDirs {
		cfg := fsutil.ScanConfig{
			Roots:     paths,
			MatchName: d,
			Workers:   4,
		}

		resChan, errChan := fsutil.ConcurrentScan(ctx, cfg)
		
		var items []fsutil.ScanResult
		var totalSize int64
		
		done := false
		for !done {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case err, ok := <-errChan:
				if ok && err != nil {
					// log warning
					fmt.Printf("Scan error: %v\n", err)
				}
			case res, ok := <-resChan:
				if !ok {
					done = true
				} else {
					if res.Size == -1 {
						// Compute directory size asynchronously or synchronously.
						// For purge, users usually want to see the size.
						ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
						size, _ := fsutil.GetDirectorySize(ctxTimeout, res.Path)
						cancel()
						res.Size = size
					}
					items = append(items, res)
					totalSize += res.Size
				}
			}
		}
		
		if len(items) > 0 {
			targetName := dirToTarget[d]
			existing := results[targetName]
			existing.Items = append(existing.Items, items...)
			existing.TotalSize += totalSize
			results[targetName] = existing
		}
	}
	
	return results, nil
}
