package ai

import (
	"strings"
)

func CleanAIOutput(raw string) string {
	raw = removeCodeBlocks(raw)
	raw = removeExplanations(raw)
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.TrimSpace(raw)
	raw = fixCommonMistakes(raw)

	return raw
}

func removeCodeBlocks(text string) string {
	text = strings.ReplaceAll(text, "```struct", "")
	text = strings.ReplaceAll(text, "```plaintext", "")
	text = strings.ReplaceAll(text, "```text", "")
	text = strings.ReplaceAll(text, "```", "")

	return text
}

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

func fixCommonMistakes(text string) string {
	lines := strings.Split(text, "\n")
	var fixed []string
	var rootFound bool

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		isRoot := !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, "  ")

		if !rootFound && isRoot {
			rootFound = true
			if !strings.HasSuffix(trimmed, "/") {
				trimmed += "/"
			}
			fixed = append(fixed, trimmed)
			continue
		}

		if rootFound && isRoot && i > 0 {
			line = "\t" + line
		}

		fixed = append(fixed, line)
	}

	return strings.Join(fixed, "\n")
}

func DetectAndWrapSingleRoot(text string, defaultRootName string) string {
	lines := strings.Split(text, "\n")

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

	if rootCount != 1 {
		var wrapped []string
		wrapped = append(wrapped, defaultRootName+"/")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				wrapped = append(wrapped, "\t"+line)
			}
		}
		return strings.Join(wrapped, "\n")
	}

	return text
}
