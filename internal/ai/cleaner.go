package ai

import (
	"strings"
)

// CleanAIOutput removes common AI artifacts and normalizes output
func CleanAIOutput(raw string) string {
	// Step 1: Remove markdown code blocks
	raw = removeCodeBlocks(raw)

	// Step 2: Remove explanatory text (common AI behavior)
	raw = removeExplanations(raw)

	// Step 3: Normalize line endings
	raw = strings.ReplaceAll(raw, "\r\n", "\n")

	// Step 4: Remove empty lines at start/end
	raw = strings.TrimSpace(raw)

	// Step 5: Fix common AI mistakes
	raw = fixCommonMistakes(raw)

	return raw
}

// removeCodeBlocks removes markdown code block markers
func removeCodeBlocks(text string) string {
	// Remove ```struct, ```plaintext, ```
	text = strings.ReplaceAll(text, "```struct", "")
	text = strings.ReplaceAll(text, "```plaintext", "")
	text = strings.ReplaceAll(text, "```text", "")
	text = strings.ReplaceAll(text, "```", "")

	return text
}

// removeExplanations removes common AI explanation patterns
func removeExplanations(text string) string {
	lines := strings.Split(text, "\n")
	var cleaned []string

	skipPrefixes := []string{
		"Here is",
		"Here's",
		"This is",
		"The structure",
		"I've created",
		"I've generated",
		"Based on",
		"Note:",
		"Important:",
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip explanation lines
		shouldSkip := false
		for _, prefix := range skipPrefixes {
			if strings.HasPrefix(trimmed, prefix) {
				shouldSkip = true
				break
			}
		}

		if !shouldSkip && trimmed != "" {
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, "\n")
}

// fixCommonMistakes fixes typical AI formatting errors
func fixCommonMistakes(text string) string {
	lines := strings.Split(text, "\n")
	var fixed []string
	var rootFound bool

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Detect if this is root level (no indentation)
		isRoot := !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, "  ")

		// First non-empty line should be root
		if !rootFound && isRoot {
			rootFound = true
			// Ensure root has trailing slash
			if !strings.HasSuffix(trimmed, "/") {
				trimmed += "/"
			}
			fixed = append(fixed, trimmed)
			continue
		}

		// If we found root, all other roots should be indented
		if rootFound && isRoot && i > 0 {
			// Convert to child of root (add tab)
			line = "\t" + line
		}

		fixed = append(fixed, line)
	}

	return strings.Join(fixed, "\n")
}

// DetectAndWrapSingleRoot wraps content in a root folder if missing
func DetectAndWrapSingleRoot(text string, defaultRootName string) string {
	lines := strings.Split(text, "\n")

	// Count root-level items
	rootCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		isRoot := !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, "  ")
		if isRoot {
			rootCount++
		}
	}

	// If multiple roots or no root with slash, wrap everything
	if rootCount != 1 {
		var wrapped []string
		wrapped = append(wrapped, defaultRootName+"/")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				// Add tab indentation
				wrapped = append(wrapped, "\t"+line)
			}
		}
		return strings.Join(wrapped, "\n")
	}

	return text
}
