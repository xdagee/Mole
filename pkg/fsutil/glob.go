package fsutil

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// ScanConfig defines the parameters for a directory scan.
type ScanConfig struct {
	Roots     []string
	MatchName string // e.g., "node_modules", ".DS_Store"
	Workers   int
}

// ScanResult holds the outcome of a scan.
type ScanResult struct {
	Path string
	Size int64
}

// ConcurrentScan walks multiple roots concurrently looking for directories/files matching MatchName.
// It explicitly skips symlinks and NTFS junction points to prevent cyclic recursion.
func ConcurrentScan(ctx context.Context, cfg ScanConfig) (<-chan ScanResult, <-chan error) {
	results := make(chan ScanResult, 100)
	errs := make(chan error, 1)

	if cfg.Workers <= 0 {
		cfg.Workers = 4
	}

	go func() {
		defer close(results)
		defer close(errs)

		var wg sync.WaitGroup
		sem := make(chan struct{}, cfg.Workers)
		
		var totalScanned uint64

		for _, root := range cfg.Roots {
			rootPath := root
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				
				err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return nil // Skip paths we can't access
					}
					
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
					}

					atomic.AddUint64(&totalScanned, 1)

					// Skip symlinks and irregular files (junctions)
					if d.Type()&os.ModeSymlink != 0 || d.Type()&os.ModeIrregular != 0 {
						if d.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}

					if d.Name() == cfg.MatchName {
						// Found a match
						info, err := d.Info()
						size := int64(0)
						if err == nil {
							if d.IsDir() {
								// We don't recursively compute size here for performance.
								// The caller can do it if needed, or we just return the match.
								// But let's return -1 to indicate directory.
								size = -1
							} else {
								size = info.Size()
							}
						}
						
						select {
						case results <- ScanResult{Path: path, Size: size}:
						case <-ctx.Done():
							return ctx.Err()
						}

						if d.IsDir() {
							return filepath.SkipDir // Don't scan inside the matched directory
						}
					}

					return nil
				})

				if err != nil && err != context.Canceled {
					select {
					case errs <- err:
					default:
					}
				}
			}()
		}

		wg.Wait()
	}()

	return results, errs
}

// GetDirectorySize computes the total size of a directory concurrently.
func GetDirectorySize(ctx context.Context, dirPath string) (int64, error) {
	var totalSize int64
	
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		if d.Type()&os.ModeSymlink != 0 || d.Type()&os.ModeIrregular != 0 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				atomic.AddInt64(&totalSize, info.Size())
			}
		}
		return nil
	})
	
	return atomic.LoadInt64(&totalSize), err
}
