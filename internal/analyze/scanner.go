package analyze

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Node represents a directory or file in the disk usage tree.
type Node struct {
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Size     int64   `json:"size"`
	IsDir    bool    `json:"is_dir"`
	Children []*Node `json:"children,omitempty"`
}

// ScanDirectory walks a directory recursively and builds a size tree.
func ScanDirectory(ctx context.Context, root string) (*Node, error) {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	info, err := os.Lstat(rootPath)
	if err != nil {
		return nil, err
	}

	rootNode := &Node{
		Name:  filepath.Base(rootPath),
		Path:  rootPath,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		rootNode.Size = info.Size()
		return rootNode, nil
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrency to avoid too many open files
	scanDir(ctx, rootNode, &wg, sem)
	wg.Wait()

	// Sort children by size descending
	sortChildren(rootNode)

	return rootNode, nil
}

func scanDir(ctx context.Context, node *Node, wg *sync.WaitGroup, sem chan struct{}) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	entries, err := os.ReadDir(node.Path)
	if err != nil {
		return
	}

	var mu sync.Mutex
	var totalSize int64

	for _, entry := range entries {
		childPath := filepath.Join(node.Path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		childNode := &Node{
			Name:  entry.Name(),
			Path:  childPath,
			IsDir: entry.IsDir(),
		}

		if entry.IsDir() {
			// Skip symlinks to avoid infinite recursion
			if info.Mode()&os.ModeSymlink != 0 {
				continue
			}
			// Fold certain hidden folders
			if strings.HasPrefix(entry.Name(), ".") && entry.Name() != ".git" {
				// We still traverse, but could optimize by folding.
				// For now, traverse normally.
			}

			wg.Add(1)
			go func(c *Node) {
				defer wg.Done()
				sem <- struct{}{}
				scanDir(ctx, c, wg, sem)
				<-sem

				mu.Lock()
				totalSize += c.Size
				node.Children = append(node.Children, c)
				mu.Unlock()
			}(childNode)
		} else {
			childNode.Size = info.Size()
			mu.Lock()
			totalSize += childNode.Size
			node.Children = append(node.Children, childNode)
			mu.Unlock()
		}
	}

	mu.Lock()
	node.Size = totalSize
	mu.Unlock()
}

func sortChildren(node *Node) {
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})
	for _, c := range node.Children {
		if c.IsDir {
			sortChildren(c)
		}
	}
}
