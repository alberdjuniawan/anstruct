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

// Parse: blueprint .struct → Tree (dari file)
func (p *Parser) Parse(ctx context.Context, blueprintPath string) (*core.Tree, error) {
	f, err := os.Open(blueprintPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Gunakan nama file tanpa ekstensi sebagai root name
	baseName := filepath.Base(blueprintPath)
	rootName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	return parseScanner(scanner, rootName)
}

// ParseString: blueprint .struct → Tree (dari string langsung)
func (p *Parser) ParseString(ctx context.Context, content string) (*core.Tree, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	// root name generik untuk AI-generated blueprints
	return parseScanner(scanner, "project")
}

// Write: Tree → blueprint .struct
func (p *Parser) Write(ctx context.Context, tree *core.Tree, path string) error {
	// Pastikan direktori parent ada
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	var b strings.Builder
	walk(tree.Root, 0, func(n *core.Node, depth int) {
		if depth > 0 { // skip root dir
			// Gunakan tab untuk indentasi (konsisten)
			b.WriteString(strings.Repeat("\t", depth-1))

			// Jika OriginalName tersedia, tulis persis itu (berguna untuk trailing slash)
			if n.OriginalName != "" {
				b.WriteString(n.OriginalName)
			} else {
				// format default: tambahkan slash untuk direktori
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

// --- helpers ---

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

		// Skip baris kosong dan komentar
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		depth := countIndent(line)
		entry := trimmed

		// Detect trailing slash -> explicit directory
		explicitDir := strings.HasSuffix(entry, "/")

		// Sanitasi nama: jika explicitDir, hapus trailing slash sebelum sanitize
		tmp := entry
		if explicitDir {
			tmp = strings.TrimSuffix(tmp, "/")
		}

		// Deteksi tipe: file jika ada titik di nama (ekstensi)
		isFileByExt := strings.Contains(tmp, ".")

		// Sanitasi nama (bersihkan path traversal atau separator)
		name := sanitize(tmp)
		if name == "" || name == "_" {
			return nil, fmt.Errorf("invalid entry name at line %d: %q", lineNum, entry)
		}

		// Create node, default directory (we'll convert to file if detected)
		n := &core.Node{
			Type:         core.NodeDir,
			Name:         name,
			OriginalName: entry, // preserve original including trailing slash if any
			Content:      "",
		}
		// If clearly file by extension and not explicitly marked as directory, mark as file
		if isFileByExt && !explicitDir {
			n.Type = core.NodeFile
		}

		// Validasi: depth tidak boleh loncat lebih dari 1 level
		parentDepth := stack[len(stack)-1].depth
		if depth > parentDepth+1 {
			return nil, fmt.Errorf("invalid indentation at line %d: jumped from depth %d to %d",
				lineNum, parentDepth, depth)
		}

		// Unwind stack sesuai depth
		for len(stack) > 0 && stack[len(stack)-1].depth >= depth {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			return nil, fmt.Errorf("stack underflow at line %d", lineNum)
		}

		parent := stack[len(stack)-1].node

		// If parent previously thought to be file, but now gets a child -> convert to dir
		if parent.Type != core.NodeDir {
			parent.Type = core.NodeDir
			// Ensure OriginalName for parent shows as directory (append slash if missing)
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

	// Post-process: ensure nodes that have children are marked as directories
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

// countIndent: fleksibel, bisa tab atau spasi
func countIndent(s string) int {
	count := 0
	spaces := 0

	for _, r := range s {
		if r == '\t' {
			count++
			spaces = 0 // reset space counter
		} else if r == ' ' {
			spaces++
			if spaces == 2 { // 2 spasi = 1 indent
				count++
				spaces = 0
			}
		} else {
			break // stop at first non-whitespace
		}
	}

	// Warning hanya sekali per proses
	if spaces > 0 && !warnedSpaces {
		fmt.Println("⚠️  Warning: indentasi pakai spasi, disarankan pakai tab untuk konsistensi")
		warnedSpaces = true
	}

	return count
}

// sanitize: bersihkan nama untuk keamanan
func sanitize(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	// Cegah path traversal
	clean := filepath.Clean(name)
	if strings.Contains(clean, "..") || filepath.IsAbs(clean) {
		// Replace berbahaya
		name = strings.ReplaceAll(name, "..", "__")
	}

	// Cegah path separator (seharusnya tidak ada / atau \)
	// Note: trailing slash already removed earlier, jadi ini aman
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		name = strings.ReplaceAll(name, "/", "-")
		name = strings.ReplaceAll(name, "\\", "-")
	}

	return name
}

// walk: traverse tree dengan callback
func walk(n *core.Node, depth int, fn func(*core.Node, int)) {
	fn(n, depth)
	for _, c := range n.Children {
		walk(c, depth+1, fn)
	}
}
