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

// Validate validates tree and removes reserved names with warnings
func (v *Validator) Validate(ctx context.Context, tree *core.Tree) error {
	return v.ValidateWithOptions(ctx, tree, false)
}

// ValidateWithOptions validates tree with configurable reserved name handling
func (v *Validator) ValidateWithOptions(ctx context.Context, tree *core.Tree, allowReserved bool) error {
	seen := map[string]bool{}
	var err error
	var skipped []string

	// First pass: collect reserved names to skip
	walk(tree.Root, "", func(path string, n *core.Node) {
		if isReserved(n.Name) {
			skipped = append(skipped, n.Name)
		}
	})

	// Remove reserved nodes from tree
	if len(skipped) > 0 {
		cleanReservedNodes(tree.Root)

		// Print warning
		fmt.Println("\n⚠️  Reserved names detected and skipped:")
		for _, name := range skipped {
			fmt.Printf("   ⏭️  %s (managed by package manager/git)\n", name)
		}
		fmt.Println("💡 These folders are typically auto-generated and shouldn't be in blueprints.\n")
	}

	// Second pass: validate remaining nodes
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

// cleanReservedNodes removes reserved children from tree
func cleanReservedNodes(n *core.Node) {
	if n == nil {
		return
	}

	// Filter out reserved children
	filtered := make([]*core.Node, 0, len(n.Children))
	for _, child := range n.Children {
		if !isReserved(child.Name) {
			filtered = append(filtered, child)
			// Recursively clean children
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
	// Reserved directories that are auto-generated
	// These should NOT be in blueprints as they're managed by tools
	reserved := []string{
		".git",         // Git repository
		"node_modules", // npm/yarn packages
		"vendor",       // PHP Composer/Go modules
		".next",        // Next.js build
		".nuxt",        // Nuxt.js build
		"dist",         // Build output
		"build",        // Build output
		".cache",       // Cache directories
		"__pycache__",  // Python cache
		".venv",        // Python virtual env
		"venv",         // Python virtual env
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
