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

func (p *Parser) Parse(ctx context.Context, blueprintPath string) (*core.Tree, error) {
	f, err := os.Open(blueprintPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	baseName := filepath.Base(blueprintPath)
	rootName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	return parseScanner(scanner, rootName)
}

func (p *Parser) ParseString(ctx context.Context, content string) (*core.Tree, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	return parseScanner(scanner, "project")
}

func (p *Parser) Write(ctx context.Context, tree *core.Tree, path string) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	var b strings.Builder
	walk(tree.Root, 0, func(n *core.Node, depth int) {
		if depth > 0 {
			b.WriteString(strings.Repeat("\t", depth-1))

			if n.OriginalName != "" {
				b.WriteString(n.OriginalName)
			} else {
				if n.Type == core.NodeDir {
					b.WriteString(n.Name + "/")
				} else {
					b.WriteString(n.Name)
				}
			}
			b.WriteString("\n")
		}
	})

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

var warnedSpaces bool

func parseScanner(scanner *bufio.Scanner, rootName string) (*core.Tree, error) {
	type frame struct {
		node  *core.Node
		depth int
	}

	root := &core.Node{
		Type:         core.NodeDir,
		Name:         rootName,
		OriginalName: rootName + "/",
	}
	stack := []frame{{node: root, depth: -1}}
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		depth := countIndent(line)
		entry := trimmed

		// 🔥 CRITICAL: Explicit folder detection
		// ONLY entries ending with "/" are folders
		explicitDir := strings.HasSuffix(entry, "/")

		tmp := entry
		if explicitDir {
			tmp = strings.TrimSuffix(tmp, "/")
		}

		// 🔥 CRITICAL: File detection by extension
		// If has dot AND not explicitly marked as folder, it's a file
		isFileByExt := strings.Contains(tmp, ".")

		name := sanitize(tmp)
		if name == "" || name == "_" {
			return nil, fmt.Errorf("invalid entry name at line %d: %q", lineNum, entry)
		}

		// 🔥 DEFAULT TO DIRECTORY first
		n := &core.Node{
			Type:         core.NodeDir,
			Name:         name,
			OriginalName: entry,
			Content:      "",
		}

		// 🔥 ONLY mark as file if:
		// 1. Has extension (contains dot)
		// 2. NOT explicitly marked as folder (no trailing /)
		if isFileByExt && !explicitDir {
			n.Type = core.NodeFile
		}

		parentDepth := stack[len(stack)-1].depth
		if depth > parentDepth+1 {
			return nil, fmt.Errorf("invalid indentation at line %d: jumped from depth %d to %d",
				lineNum, parentDepth, depth)
		}

		for len(stack) > 0 && stack[len(stack)-1].depth >= depth {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			return nil, fmt.Errorf("stack underflow at line %d", lineNum)
		}

		parent := stack[len(stack)-1].node

		// If parent has children, it must be a directory
		if parent.Type != core.NodeDir {
			parent.Type = core.NodeDir
			if !strings.HasSuffix(parent.OriginalName, "/") {
				parent.OriginalName = parent.Name + "/"
			}
		}

		parent.Children = append(parent.Children, n)
		stack = append(stack, frame{node: n, depth: depth})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	// Post-processing: fix any nodes with children to be directories
	var fix func(*core.Node)
	fix = func(n *core.Node) {
		if len(n.Children) > 0 {
			n.Type = core.NodeDir
			if n.OriginalName == "" || !strings.HasSuffix(n.OriginalName, "/") {
				n.OriginalName = n.Name + "/"
			}
		}
		for _, c := range n.Children {
			fix(c)
		}
	}
	fix(root)

	return &core.Tree{Root: root}, nil
}

func countIndent(s string) int {
	count := 0
	spaces := 0

	for _, r := range s {
		if r == '\t' {
			count++
			spaces = 0
		} else if r == ' ' {
			spaces++
			if spaces == 2 {
				count++
				spaces = 0
			}
		} else {
			break
		}
	}

	if spaces > 0 && !warnedSpaces {
		fmt.Println("⚠️  Warning: indentasi pakai spasi, disarankan pakai tab untuk konsistensi")
		warnedSpaces = true
	}

	return count
}

func sanitize(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	clean := filepath.Clean(name)
	if strings.Contains(clean, "..") || filepath.IsAbs(clean) {
		name = strings.ReplaceAll(name, "..", "__")
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		name = strings.ReplaceAll(name, "/", "-")
		name = strings.ReplaceAll(name, "\\", "-")
	}

	return name
}

func walk(n *core.Node, depth int, fn func(*core.Node, int)) {
	fn(n, depth)
	for _, c := range n.Children {
		walk(c, depth+1, fn)
	}
}
