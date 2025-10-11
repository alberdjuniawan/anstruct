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
	root := &core.Node{Type: core.NodeDir, Name: filepath.Base(inputDir), OriginalName: filepath.Base(inputDir) + "/"}

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

		// find existing child with same name
		for _, c := range cur.Children {
			if c.Name == name {
				next = c
				break
			}
		}
		if next == nil {
			t := core.NodeDir
			orig := name + "/"
			if last && !d.IsDir() {
				t = core.NodeFile
				orig = name
			}
			next = &core.Node{Type: t, Name: name, OriginalName: orig}
			cur.Children = append(cur.Children, next)
		} else {
			// If existing child was created as file earlier but now we realize it's a dir (because deeper path exists)
			if last == false && next.Type == core.NodeFile {
				next.Type = core.NodeDir
				if !strings.HasSuffix(next.OriginalName, "/") {
					next.OriginalName = next.Name + "/"
				}
			}
		}
		cur = next
	}
}
