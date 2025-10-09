package parser

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alberdjuniawan/anstruct/internal/core"
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
	root := &core.Node{
		Type:         core.NodeDir,
		Name:         filepath.Base(strings.TrimSuffix(blueprintPath, filepath.Ext(blueprintPath))),
		OriginalName: filepath.Base(strings.TrimSuffix(blueprintPath, filepath.Ext(blueprintPath))),
	}
	stack := []frame{{node: root, depth: -1}}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue // skip kosong & komentar
		}

		depth := countIndent(line)
		entry := strings.TrimSpace(line)
		isFile := strings.Contains(entry, ".")
		name := sanitize(entry)

		n := &core.Node{
			Type:         core.NodeDir,
			Name:         name,
			OriginalName: entry,
			Content:      "",
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

var warnedSpaces bool

// countIndent: fleksibel, bisa tab atau spasi
func countIndent(s string) int {
	count := 0
	spaces := 0

	for _, r := range s {
		if r == '\t' {
			count++
			spaces = 0
		} else if r == ' ' {
			spaces++
			if spaces == 2 { // anggap 2 spasi = 1 indent
				count++
				spaces = 0
			}
			if !warnedSpaces {
				fmt.Println("⚠️  Warning: indentasi pakai spasi, disarankan pakai tab untuk konsistensi")
				warnedSpaces = true
			}
		} else {
			break
		}
	}
	return count
}

func sanitize(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "_"
	}
	// cegah traversal
	clean := filepath.Clean(name)
	if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return strings.ReplaceAll(name, "..", "_")
	}
	// ganti path separator
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	return name
}

func walk(n *core.Node, depth int, fn func(*core.Node, int)) {
	fn(n, depth)
	for _, c := range n.Children {
		walk(c, depth+1, fn)
	}
}
