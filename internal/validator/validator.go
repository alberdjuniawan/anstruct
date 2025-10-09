package validator

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/alberdjuniawan/anstruct/internal/core"
)

type Validator struct{}

func New() *Validator { return &Validator{} }

func (v *Validator) Validate(ctx context.Context, tree *core.Tree) error {
	seen := map[string]bool{}
	var err error

	walk(tree.Root, "", func(path string, n *core.Node) {
		if seen[path] {
			err = errors.New("duplicate path: " + path)
			return
		}
		seen[path] = true

		if isReserved(n.Name) {
			err = errors.New("reserved name: " + n.Name)
			return
		}

		// check traversal using OriginalName (raw name from parser)
		if isTraversal(n.OriginalName) {
			err = errors.New("path traversal detected: " + n.OriginalName)
			return
		}
	})
	return err
}

// --- helpers ---

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
	reserved := []string{".git", "node_modules", "vendor"}
	if strings.HasSuffix(strings.ToLower(name), ".struct") {
		return true
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
