package reverser

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/alberdjuniawan/anstruct/internal/core"
)

type Reverser struct{}

func New() *Reverser { return &Reverser{} }

func (r *Reverser) Reverse(ctx context.Context, inputDir string) (*core.Tree, error) {
	root := &core.Node{Type: core.NodeDir, Name: filepath.Base(inputDir)}

	err := filepath.WalkDir(inputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == inputDir {
			return nil // skip root
		}

		rel, _ := filepath.Rel(inputDir, path)
		parts := strings.Split(rel, string(os.PathSeparator))
		insert(root, parts, d)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &core.Tree{Root: root}, nil
}

func insert(root *core.Node, parts []string, d os.DirEntry) {
	cur := root
	for i, name := range parts {
		last := i == len(parts)-1
		var next *core.Node

		for _, c := range cur.Children {
			if c.Name == name {
				next = c
				break
			}
		}
		if next == nil {
			t := core.NodeDir
			if last && !d.IsDir() {
				t = core.NodeFile
			}
			next = &core.Node{Type: t, Name: name}
			cur.Children = append(cur.Children, next)
		}
		cur = next
	}
}
