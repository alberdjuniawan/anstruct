package validator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alberdjuniawan/anstruct/internal/core"
)

type Validator struct{}

func New() *Validator { return &Validator{} }

func (v *Validator) Validate(ctx context.Context, tree *core.Tree) error {
	return v.ValidateWithOptions(ctx, tree, false)
}

func (v *Validator) ValidateWithOptions(ctx context.Context, tree *core.Tree, allowReserved bool) error {
	seen := map[string]bool{}
	var err error
	var skipped []string

	walk(tree.Root, "", func(path string, n *core.Node) {
		if isReserved(n.Name) {
			skipped = append(skipped, n.Name)
		}
	})

	if len(skipped) > 0 {
		cleanReservedNodes(tree.Root)

		fmt.Println("\n⚠️  Reserved names detected and skipped:")
		for _, name := range skipped {
			fmt.Printf("   ⏭️  %s (managed by package manager/git)\n", name)
		}
		fmt.Println("💡 These folders are typically auto-generated and shouldn't be in blueprints.\n")
	}

	walk(tree.Root, "", func(path string, n *core.Node) {
		if seen[path] {
			err = errors.New("duplicate path: " + path)
			return
		}
		seen[path] = true

		if isTraversal(n.OriginalName) {
			err = errors.New("path traversal detected: " + n.OriginalName)
			return
		}
	})

	return err
}

func cleanReservedNodes(n *core.Node) {
	if n == nil {
		return
	}

	filtered := make([]*core.Node, 0, len(n.Children))
	for _, child := range n.Children {
		if !isReserved(child.Name) {
			filtered = append(filtered, child)
			cleanReservedNodes(child)
		}
	}
	n.Children = filtered
}

func walk(n *core.Node, prefix string, fn func(path string, n *core.Node)) {
	path := prefix
	if prefix == "" {
		path = n.Name
	} else {
		path = prefix + "/" + n.Name
	}
	fn(path, n)
	for _, c := range n.Children {
		walk(c, path, fn)
	}
}

func isReserved(name string) bool {
	reserved := []string{
		".git",
		"node_modules",
		"vendor",
		".next",
		".nuxt",
		"dist",
		"build",
		".cache",
		"__pycache__",
		".venv",
		"venv",
	}

	for _, r := range reserved {
		if strings.EqualFold(name, r) {
			return true
		}
	}
	return false
}

func isTraversal(raw string) bool {
	if raw == "" {
		return true
	}
	clean := filepath.Clean(raw)
	return strings.HasPrefix(clean, "..") || filepath.IsAbs(clean)
}
