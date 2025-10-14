package converter

import (
	"context"
	"fmt"
	"strings"

	"github.com/alberdjuniawan/anstruct/internal/ai"
	"github.com/alberdjuniawan/anstruct/internal/core"
)

type Converter struct {
	Normalizer *Normalizer
}

func New() *Converter {
	return &Converter{
		Normalizer: NewNormalizer(nil, nil, ModeManual),
	}
}

func NewWithAI(provider ai.Provider, parser core.Parser, mode NormalizationMode) *Converter {
	return &Converter{
		Normalizer: NewNormalizer(provider, parser, mode),
	}
}

func (c *Converter) Convert(ctx context.Context, input string) (*core.Tree, DetectedFormat, error) {
	if c.Normalizer.Provider != nil {
		return c.convertWithAI(ctx, input)
	}

	return c.convertManual(ctx, input)
}

func (c *Converter) convertWithAI(ctx context.Context, input string) (*core.Tree, DetectedFormat, error) {
	quality := c.Normalizer.DetectQuality(input)
	format := c.DetectFormat(input)

	fmt.Printf("📊 Input quality score: %d/100\n", quality)
	fmt.Printf("🔍 Detected format: %s\n", format)

	suggestedMode := c.Normalizer.SuggestMode(input)
	if c.Normalizer.Mode == ModeAuto {
		fmt.Printf("💡 Suggested mode: %s\n", suggestedMode)
	}

	normalized, err := c.Normalizer.Normalize(ctx, input)
	if err != nil {
		return nil, format, fmt.Errorf("normalization failed: %w", err)
	}

	parser := c.Normalizer.Parser
	if parser == nil {
		return c.parseNormalized(normalized, format)
	}

	tree, err := parser.ParseString(ctx, normalized)
	if err != nil {
		return nil, format, fmt.Errorf("failed to parse normalized output: %w", err)
	}

	return tree, format, nil
}

func (c *Converter) convertManual(ctx context.Context, input string) (*core.Tree, DetectedFormat, error) {
	format := c.DetectFormat(input)

	switch format {
	case FormatTree:
		return c.convertTreeFormat(input)
	case FormatLs:
		return c.convertLsFormat(input)
	case FormatMarkdown:
		return c.convertMarkdownFormat(input)
	case FormatPlain:
		return c.convertPlainFormat(input)
	default:
		return nil, format, fmt.Errorf("unsupported format: %s", format)
	}
}

type DetectedFormat string

const (
	FormatTree     DetectedFormat = "tree"
	FormatLs       DetectedFormat = "ls"
	FormatMarkdown DetectedFormat = "markdown"
	FormatPlain    DetectedFormat = "plain"
	FormatJSON     DetectedFormat = "json"
	FormatUnknown  DetectedFormat = "unknown"
)

func (c *Converter) DetectFormat(input string) DetectedFormat {
	input = strings.TrimSpace(input)

	if strings.Contains(input, "├──") || strings.Contains(input, "└──") {
		return FormatTree
	}
	if strings.Contains(input, ".:") || strings.Contains(input, "./") {
		return FormatLs
	}
	if strings.Contains(input, "```") {
		return FormatMarkdown
	}
	if strings.HasPrefix(input, "{") || strings.HasPrefix(input, "[") {
		return FormatJSON
	}

	return FormatPlain
}

func (c *Converter) convertTreeFormat(input string) (*core.Tree, DetectedFormat, error) {
	lines := strings.Split(input, "\n")
	rootName := "project"

	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if firstLine != "" {
			rootName = strings.TrimSuffix(firstLine, "/")
		}
		lines = lines[1:]
	}

	root := &core.Node{
		Type:         core.NodeDir,
		Name:         rootName,
		OriginalName: rootName + "/",
	}

	var cleaned []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		cleanedLine := removeTreeSymbols(line)
		if cleanedLine == "" {
			continue
		}

		depth := calculateTreeDepth(line)
		tabLine := strings.Repeat("\t", depth) + cleanedLine
		cleaned = append(cleaned, tabLine)
	}

	return c.parseCleanedLines(root, cleaned)
}

func (c *Converter) convertMarkdownFormat(input string) (*core.Tree, DetectedFormat, error) {
	input = strings.ReplaceAll(input, "```", "")
	return c.convertTreeFormat(input)
}

func (c *Converter) convertLsFormat(input string) (*core.Tree, DetectedFormat, error) {
	root := &core.Node{
		Type:         core.NodeDir,
		Name:         "project",
		OriginalName: "project/",
	}
	return &core.Tree{Root: root}, FormatLs, nil
}

func (c *Converter) convertPlainFormat(input string) (*core.Tree, DetectedFormat, error) {
	lines := strings.Split(input, "\n")
	rootName := "project"

	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if firstLine != "" {
			rootName = strings.TrimSuffix(firstLine, "/")
		}
	}

	root := &core.Node{
		Type:         core.NodeDir,
		Name:         rootName,
		OriginalName: rootName + "/",
	}

	return c.parseCleanedLines(root, lines)
}

func (c *Converter) parseNormalized(normalized string, format DetectedFormat) (*core.Tree, DetectedFormat, error) {
	lines := strings.Split(normalized, "\n")
	rootName := "project"

	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if firstLine != "" {
			rootName = strings.TrimSuffix(firstLine, "/")
		}
	}

	root := &core.Node{
		Type:         core.NodeDir,
		Name:         rootName,
		OriginalName: rootName + "/",
	}

	return c.parseCleanedLines(root, lines)
}

func (c *Converter) parseCleanedLines(root *core.Node, lines []string) (*core.Tree, DetectedFormat, error) {
	type frame struct {
		node  *core.Node
		depth int
	}

	stack := []frame{{node: root, depth: -1}}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		depth := 0
		for _, ch := range line {
			if ch == '\t' {
				depth++
			} else {
				break
			}
		}

		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}

		isDir := strings.HasSuffix(name, "/")
		if isDir {
			name = strings.TrimSuffix(name, "/")
		}

		nodeType := core.NodeFile
		originalName := name
		if isDir || !strings.Contains(name, ".") {
			nodeType = core.NodeDir
			originalName = name + "/"
		}

		n := &core.Node{
			Type:         nodeType,
			Name:         name,
			OriginalName: originalName,
		}

		for len(stack) > 0 && stack[len(stack)-1].depth >= depth {
			stack = stack[:len(stack)-1]
		}

		if len(stack) > 0 {
			parent := stack[len(stack)-1].node
			parent.Children = append(parent.Children, n)
			stack = append(stack, frame{node: n, depth: depth})
		}
	}

	return &core.Tree{Root: root}, FormatPlain, nil
}

func (c *Converter) ConvertToString(tree *core.Tree) string {
	var b strings.Builder

	var walk func(*core.Node, int)
	walk = func(n *core.Node, depth int) {
		if depth > 0 {
			b.WriteString(strings.Repeat("\t", depth-1))
			if n.Type == core.NodeDir {
				b.WriteString(n.Name + "/")
			} else {
				b.WriteString(n.Name)
			}
			b.WriteString("\n")
		}
		for _, child := range n.Children {
			walk(child, depth+1)
		}
	}

	walk(tree.Root, 0)
	return b.String()
}

func removeTreeSymbols(line string) string {
	symbols := []string{"├──", "└──", "│", "├─", "└─", "─"}
	for _, sym := range symbols {
		line = strings.ReplaceAll(line, sym, "")
	}
	return strings.TrimSpace(line)
}

func calculateTreeDepth(line string) int {
	depth := 0
	for _, ch := range line {
		if ch == ' ' || ch == '│' {
			depth++
		} else if ch == '├' || ch == '└' {
			break
		} else if ch != '─' {
			break
		}
	}
	return depth / 4
}
