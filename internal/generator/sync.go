package generator

import (
	"os"
	"path/filepath"

	"github.com/alberdjuniawan/anstruct/internal/core"
)

func CollectAllowed(n *core.Node, prefix string, allowed map[string]bool) {
	var path string
	if prefix == "" {
		path = n.Name
	} else {
		path = filepath.Join(prefix, n.Name)
	}
	allowed[path] = true
	for _, c := range n.Children {
		CollectAllowed(c, path, allowed)
	}
}

func CleanupExtra(outputDir string, allowed map[string]bool) error {
	return filepath.WalkDir(outputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == outputDir {
			return nil
		}
		rel, _ := filepath.Rel(outputDir, path)
		if !allowed[rel] {
			if d.IsDir() {
				return os.RemoveAll(path)
			}
			return os.Remove(path)
		}
		return nil
	})
}
