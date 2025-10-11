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

func ValidateStructOutput(output string) error {
	if strings.TrimSpace(output) == "" {
		return ErrInvalidFormat
	}

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

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		hasContent = true

		if strings.HasPrefix(line, "\t") {
			usesTab = true
		} else if strings.HasPrefix(line, "  ") {
			usesSpace = true
		}

		if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") {
			rootCount++
			if rootCount > 1 {
				return fmt.Errorf("%w at line %d: '%s' (only one root folder allowed)",
					ErrMultipleRoots, lineNum, trimmed)
			}

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

	if usesTab && usesSpace {
		return fmt.Errorf("%w: mixing tabs and spaces detected", ErrInconsistentIndent)
	}

	return nil
}

func cleanMarkdown(output string) string {
	output = strings.ReplaceAll(output, "```struct", "")
	output = strings.ReplaceAll(output, "```", "")

	lines := strings.Split(output, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "struct" || trimmed == "plaintext" {
			continue
		}
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}
