package ai

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidFormat      = errors.New("invalid .struct format")
	ErrNoRootFolder       = errors.New("missing root folder")
	ErrMultipleRoots      = errors.New("multiple root folders detected")
	ErrInconsistentIndent = errors.New("inconsistent indentation")
	ErrMissingSlash       = errors.New("folder missing trailing slash")
)

// ValidateStructOutput validasi hasil AI output sebelum digunakan
func ValidateStructOutput(output string) error {
	if strings.TrimSpace(output) == "" {
		return ErrInvalidFormat
	}

	// Clean up output: remove markdown code blocks if present
	output = cleanMarkdown(output)

	scanner := bufio.NewScanner(strings.NewReader(output))
	lineNum := 0
	rootCount := 0
	usesTab := false
	usesSpace := false
	hasContent := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines dan komentar
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		hasContent = true

		// Deteksi indentasi
		if strings.HasPrefix(line, "\t") {
			usesTab = true
		} else if strings.HasPrefix(line, "  ") { // 2 spaces minimum
			usesSpace = true
		}

		// Hitung root level (tidak ada indentasi)
		if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") {
			rootCount++
			if rootCount > 1 {
				return fmt.Errorf("%w at line %d: '%s' (only one root folder allowed)",
					ErrMultipleRoots, lineNum, trimmed)
			}

			// Root must be a folder (end with /)
			if !strings.HasSuffix(trimmed, "/") {
				return fmt.Errorf("%w: root must be a folder ending with '/' at line %d: '%s'",
					ErrNoRootFolder, lineNum, trimmed)
			}
		}
	}

	if !hasContent {
		return ErrInvalidFormat
	}

	if rootCount == 0 {
		return ErrNoRootFolder
	}

	// Warning jika mixing tab dan space
	if usesTab && usesSpace {
		return fmt.Errorf("%w: mixing tabs and spaces detected", ErrInconsistentIndent)
	}

	return nil
}

// cleanMarkdown removes code block markers if AI includes them
func cleanMarkdown(output string) string {
	// Remove ```struct or ``` code blocks
	output = strings.ReplaceAll(output, "```struct", "")
	output = strings.ReplaceAll(output, "```", "")

	// Remove leading/trailing whitespace
	lines := strings.Split(output, "\n")
	var cleaned []string
	for _, line := range lines {
		// Skip lines that are just markdown artifacts
		trimmed := strings.TrimSpace(line)
		if trimmed == "struct" || trimmed == "plaintext" {
			continue
		}
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

// Removed BuildRetryPrompt - now in prompts.go for better organization
