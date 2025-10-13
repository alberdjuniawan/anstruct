package normalizer

import (
	"bufio"
	"fmt"
	"strings"
	"unicode"
)

// Normalizer provides non-AI structure normalization
type Normalizer struct{}

func New() *Normalizer {
	return &Normalizer{}
}

// NormalizeToStruct attempts to convert various formats to .struct format
// Returns normalized content and confidence score (0-100)
func (n *Normalizer) NormalizeToStruct(input string) (string, int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", 0, fmt.Errorf("empty input")
	}

	// Detect format and normalize
	if format := detectFormat(input); format != "" {
		switch format {
		case "tree":
			return normalizeTreeFormat(input)
		case "markdown":
			return normalizeMarkdown(input)
		case "indented":
			return normalizeIndented(input)
		case "struct":
			// Already in .struct format, just clean it
			return cleanStructFormat(input), 95, nil
		}
	}

	// Fallback: try generic normalization
	return genericNormalize(input)
}

// detectFormat detects the input format
func detectFormat(input string) string {
	lines := strings.Split(input, "\n")

	// Check for tree format (├──, └──, │)
	if containsTreeChars(input) {
		return "tree"
	}

	// Check for markdown (-, *, **)
	if isMarkdownFormat(lines) {
		return "markdown"
	}

	// Check if already .struct format (tabs, folders with /)
	if isStructFormat(lines) {
		return "struct"
	}

	// Default to indented format
	return "indented"
}

func containsTreeChars(s string) bool {
	treeChars := []string{"├", "└", "│", "─", "├─", "└─"}
	for _, char := range treeChars {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}

func isMarkdownFormat(lines []string) bool {
	markdownCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-") ||
			strings.HasPrefix(trimmed, "*") ||
			strings.HasPrefix(trimmed, "+") ||
			strings.Contains(trimmed, "**") {
			markdownCount++
		}
	}
	return markdownCount > len(lines)/3
}

func isStructFormat(lines []string) bool {
	hasRoot := false
	hasIndent := false

	for i, line := range lines {
		if i == 0 && strings.HasSuffix(strings.TrimSpace(line), "/") {
			hasRoot = true
		}
		if strings.HasPrefix(line, "\t") {
			hasIndent = true
		}
	}

	return hasRoot && hasIndent
}

// normalizeTreeFormat converts tree command output to .struct
func normalizeTreeFormat(input string) (string, int, error) {
	lines := strings.Split(input, "\n")
	var result []string
	var root string

	// Parse tree structure
	type item struct {
		name  string
		depth int
		isDir bool
	}
	var items []item

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// First non-empty line is usually the root
		if i == 0 || (root == "" && !containsTreeChars(line)) {
			root = strings.TrimSpace(line)
			if !strings.HasSuffix(root, "/") {
				root += "/"
			}
			continue
		}

		// Skip summary lines (e.g., "5 directories, 10 files")
		if strings.Contains(line, "directories") || strings.Contains(line, "files") {
			continue
		}

		// Remove tree characters
		cleaned := line
		for _, char := range []string{"├──", "└──", "│", "├─", "└─", "─"} {
			cleaned = strings.ReplaceAll(cleaned, char, "")
		}
		cleaned = strings.TrimLeft(cleaned, " ")

		if cleaned == "" {
			continue
		}

		// Calculate depth based on original indentation
		depth := 0
		for _, r := range line {
			if r == ' ' || r == '│' || r == '├' || r == '└' || r == '─' {
				depth++
			} else {
				break
			}
		}
		depth = depth / 4 // Approximate depth

		name := strings.TrimSpace(cleaned)
		isDir := strings.HasSuffix(name, "/") || !strings.Contains(name, ".")

		if isDir && !strings.HasSuffix(name, "/") {
			name += "/"
		}

		items = append(items, item{name: name, depth: depth, isDir: isDir})
	}

	// Build output
	if root == "" {
		root = "project/"
	}
	result = append(result, root)

	for _, it := range items {
		indent := strings.Repeat("\t", it.depth+1)
		result = append(result, indent+it.name)
	}

	output := strings.Join(result, "\n")
	return output, 80, nil
}

// normalizeMarkdown converts markdown lists to .struct
func normalizeMarkdown(input string) (string, int, error) {
	lines := strings.Split(input, "\n")
	var result []string
	root := "project/"

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and headers
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Remove markdown list markers
		cleaned := trimmed
		for _, marker := range []string{"- ", "* ", "+ ", "• "} {
			if strings.HasPrefix(cleaned, marker) {
				cleaned = strings.TrimPrefix(cleaned, marker)
				break
			}
		}

		// Remove bold markers
		cleaned = strings.ReplaceAll(cleaned, "**", "")
		cleaned = strings.TrimSpace(cleaned)

		if cleaned == "" {
			continue
		}

		// Calculate indentation level
		originalSpaces := len(line) - len(strings.TrimLeft(line, " \t"))
		depth := originalSpaces / 2
		if depth < 0 {
			depth = 0
		}

		// Determine if directory
		isDir := strings.HasSuffix(cleaned, "/") || !strings.Contains(cleaned, ".")
		if isDir && !strings.HasSuffix(cleaned, "/") {
			cleaned += "/"
		}

		if depth == 0 && len(result) == 0 {
			root = cleaned
		} else {
			indent := strings.Repeat("\t", depth)
			result = append(result, indent+cleaned)
		}
	}

	if len(result) == 0 {
		result = append(result, root)
	} else if result[0] != root {
		result = append([]string{root}, result...)
	}

	output := strings.Join(result, "\n")
	return output, 75, nil
}

// normalizeIndented converts indented text to .struct
func normalizeIndented(input string) (string, int, error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	var result []string
	var root string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Calculate indentation
		spaces := 0
		tabs := 0
		for _, r := range line {
			if r == ' ' {
				spaces++
			} else if r == '\t' {
				tabs++
			} else {
				break
			}
		}

		// Convert spaces to tab depth (assume 2 or 4 spaces = 1 tab)
		depth := tabs
		if spaces > 0 {
			depth += spaces / 4
			if spaces%4 >= 2 {
				depth++
			}
		}

		// Determine if directory
		isDir := strings.HasSuffix(trimmed, "/") || !strings.Contains(trimmed, ".")
		name := trimmed
		if isDir && !strings.HasSuffix(name, "/") {
			name += "/"
		}

		// First item is root
		if root == "" {
			root = name
			result = append(result, root)
			continue
		}

		indent := strings.Repeat("\t", depth)
		result = append(result, indent+name)
	}

	if len(result) == 0 {
		return "", 0, fmt.Errorf("no valid structure found")
	}

	output := strings.Join(result, "\n")
	return output, 70, nil
}

// cleanStructFormat cleans already .struct formatted content
func cleanStructFormat(input string) string {
	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Preserve tab indentation
		tabCount := 0
		for _, r := range line {
			if r == '\t' {
				tabCount++
			} else {
				break
			}
		}

		indent := strings.Repeat("\t", tabCount)
		result = append(result, indent+trimmed)
	}

	return strings.Join(result, "\n")
}

// genericNormalize attempts generic normalization
func genericNormalize(input string) (string, int, error) {
	lines := strings.Split(input, "\n")
	var result []string
	root := "project/"

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Remove common prefixes
		for _, prefix := range []string{"- ", "* ", "+ ", "• ", "> "} {
			trimmed = strings.TrimPrefix(trimmed, prefix)
		}

		// Calculate depth based on leading spaces
		leadingSpaces := len(line) - len(strings.TrimLeft(line, " \t"))
		depth := leadingSpaces / 2

		// Check if it's a directory
		isDir := !strings.Contains(trimmed, ".")
		if isDir && !strings.HasSuffix(trimmed, "/") {
			trimmed += "/"
		}

		if depth == 0 && len(result) == 0 {
			root = trimmed
			result = append(result, root)
		} else {
			indent := strings.Repeat("\t", depth)
			result = append(result, indent+trimmed)
		}
	}

	if len(result) == 0 {
		return "", 0, fmt.Errorf("could not parse structure")
	}

	output := strings.Join(result, "\n")
	return output, 50, nil // Low confidence for generic
}

// ValidateStructOutput validates the normalized output
func (n *Normalizer) ValidateStructOutput(output string) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("empty output")
	}

	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return fmt.Errorf("no lines in output")
	}

	// Check first line is root
	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasSuffix(firstLine, "/") {
		return fmt.Errorf("first line must be root folder ending with /")
	}

	// Check for tab indentation
	hasContent := false
	for i, line := range lines {
		if i == 0 {
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		hasContent = true

		// Must start with tab if not root
		if !strings.HasPrefix(line, "\t") {
			return fmt.Errorf("line %d must be indented with tab: %s", i+1, line)
		}
	}

	if !hasContent {
		return fmt.Errorf("no content under root folder")
	}

	return nil
}

// isLetter checks if rune is a letter (including UTF-8)
func isLetter(r rune) bool {
	return unicode.IsLetter(r)
}
