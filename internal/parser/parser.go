package parser

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourorg/anstruct/internal/core"
)

type Parser struct{}

func New() *Parser { return &Parser{} }

// Parse: blueprint .struct → Tree
func (p *Parser) Parse(ctx context.Context, blueprintPath string) (*core.Tree, error) {
	f, err := os.Open(blueprintPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	type frame struct {
		node  *core.Node
		depth int
	}
	root := &core.Node{Type: core.NodeDir, Name: filepath.Base(strings.TrimSuffix(blueprintPath, filepath.Ext(blueprintPath)))}
	stack := []frame{{node: root, depth: -1}}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue // skip kosong & komentar
		}

		depth := countTabs(line)
		entry := strings.TrimSpace(line)
		isFile := strings.Contains(entry, ".")
		name := sanitize(entry)

		n := &core.Node{
			Type:    core.NodeDir,
			Name:    name,
			Content: "",
		}
		if isFile {
			n.Type = core.NodeFile
		}

		// unwind stack sesuai depth
		for len(stack) > 0 && stack[len(stack)-1].depth >= depth {
			stack = stack[:len(stack)-1]
		}
		parent := stack[len(stack)-1].node
		parent.Children = append(parent.Children, n)
		stack = append(stack, frame{node: n, depth: depth})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &core.Tree{Root: root}, nil
}

// Write: Tree → blueprint .struct
func (p *Parser) Write(ctx context.Context, tree *core.Tree, path string) error {
	var b strings.Builder
	walk(tree.Root, 0, func(n *core.Node, depth int) {
		if depth > 0 { // skip root dir
			b.WriteString(strings.Repeat("\t", depth-1))
			b.WriteString(n.Name)
			b.WriteString("\n")
		}
	})
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// --- helpers ---

func countTabs(s string) int {
	i := 0
	for _, r := range s {
		if r == '\t' {
			i++
		} else {
			break
		}
	}
	return i
}

func sanitize(name string) string {
	// anti path traversal
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, string(os.PathSeparator), "-")
	return name
}

func walk(n *core.Node, depth int, fn func(*core.Node, int)) {
	fn(n, depth)
	for _, c := range n.Children {
		walk(c, depth+1, fn)
	}
}
