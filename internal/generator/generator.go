package generator

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/alberdjuniawan/anstruct/internal/core"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Generate(ctx context.Context, tree *core.Tree, outputDir string, opts core.GenerateOptions) (core.Receipt, error) {
	if hasTraversal(outputDir) {
		return core.Receipt{}, core.ErrPathTraversal
	}
	receipt := core.Receipt{}
	err := writeNode(tree.Root, outputDir, opts, &receipt)
	return receipt, err
}

func writeNode(n *core.Node, base string, opts core.GenerateOptions, r *core.Receipt) error {
	target := filepath.Join(base, n.Name)

	switch n.Type {
	case core.NodeDir:
		if !opts.DryRun {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		}
		r.CreatedDirs = append(r.CreatedDirs, target)
		for _, c := range n.Children {
			if err := writeNode(c, target, opts, r); err != nil {
				return err
			}
		}
	case core.NodeFile:
		if !opts.DryRun {
			flags := os.O_CREATE | os.O_WRONLY
			if opts.Force {
				flags |= os.O_TRUNC
			} else {
				if _, err := os.Stat(target); err == nil {
					return errors.New("file exists: " + target)
				}
			}
			f, err := os.OpenFile(target, flags, 0o644)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := f.WriteString(n.Content); err != nil {
				return err
			}
		}
		r.CreatedFiles = append(r.CreatedFiles, target)
	}
	return nil
}

func hasTraversal(p string) bool {
	clean := filepath.Clean(p)
	return clean == ".." || clean == "." || len(clean) == 0
}
