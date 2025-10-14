package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alberdjuniawan/anstruct/internal/core"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Generate(ctx context.Context, tree *core.Tree, outputDir string, opts core.GenerateOptions) (core.Receipt, error) {
	receipt := core.Receipt{}

	if !opts.DryRun {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return receipt, err
		}
	}

	for _, c := range tree.Root.Children {
		if err := writeNode(c, outputDir, opts, &receipt); err != nil {
			return receipt, err
		}
	}

	return receipt, nil
}

func writeNode(n *core.Node, base string, opts core.GenerateOptions, r *core.Receipt) error {
	target := filepath.Join(base, n.Name)

	switch n.Type {
	case core.NodeDir:
		if !opts.DryRun {
			if info, err := os.Stat(target); err == nil {
				if !info.IsDir() {
					return fmt.Errorf("cannot create directory %s: path exists as a file", target)
				}
			} else {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", target, err)
				}
			}
		}
		r.CreatedDirs = append(r.CreatedDirs, target)
		
		for _, c := range n.Children {
			if err := writeNode(c, target, opts, r); err != nil {
				return err
			}
		}

	case core.NodeFile:
		parentDir := filepath.Dir(target)
		if !opts.DryRun {
			if err := os.MkdirAll(parentDir, 0o755); err != nil {
				return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
			}

			if !opts.Force {
				if info, err := os.Stat(target); err == nil {
					if info.IsDir() {
						return fmt.Errorf("cannot create file %s: path exists as a directory", target)
					}
					return fmt.Errorf("file exists: %s (use --force to overwrite)", target)
				}
			} else {
				if existing, err := os.ReadFile(target); err == nil {
					if string(existing) == n.Content {
						return nil
					}
				}
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", target, err)
			}
			defer f.Close()

			if _, err := f.WriteString(n.Content); err != nil {
				return fmt.Errorf("failed to write file %s: %w", target, err)
			}
		}
		r.CreatedFiles = append(r.CreatedFiles, target)
	}
	return nil
}