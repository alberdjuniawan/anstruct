package converter

import (
	"context"
	"fmt"

	"github.com/alberdjuniawan/anstruct/internal/ai"
	"github.com/alberdjuniawan/anstruct/internal/core"
)

type NormalizationMode string

const (
	ModeAuto    NormalizationMode = "auto"
	ModeAI      NormalizationMode = "ai"
	ModeManual  NormalizationMode = "manual"
	ModeOffline NormalizationMode = "offline"
)

type Normalizer struct {
	Provider ai.Provider
	Parser   core.Parser
	Mode     NormalizationMode
}

func NewNormalizer(provider ai.Provider, parser core.Parser, mode NormalizationMode) *Normalizer {
	if mode == "" {
		mode = ModeAuto
	}
	return &Normalizer{
		Provider: provider,
		Parser:   parser,
		Mode:     mode,
	}
}

func (n *Normalizer) Normalize(ctx context.Context, messyInput string) (string, error) {
	switch n.Mode {
	case ModeAI:
		return n.normalizeWithAI(ctx, messyInput)

	case ModeManual, ModeOffline:
		return n.normalizeManual(messyInput)

	case ModeAuto:
		result, err := n.normalizeWithAI(ctx, messyInput)
		if err == nil {
			return result, nil
		}

		fmt.Println("⚠️  AI normalization failed, falling back to manual parsing...")
		return n.normalizeManual(messyInput)

	default:
		return "", fmt.Errorf("unsupported normalization mode: %s", n.Mode)
	}
}

func (n *Normalizer) normalizeWithAI(ctx context.Context, input string) (string, error) {
	if n.Provider == nil {
		return "", fmt.Errorf("AI provider not available")
	}

	prompt := ai.BuildNormalizationPrompt(input)

	normalized, err := n.Provider.GenerateBlueprint(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("AI normalization failed: %w", err)
	}

	cleaned := ai.CleanAIOutput(normalized)

	if err := ai.ValidateStructOutput(cleaned); err != nil {
		retryPrompt := ai.RetryPrompt(input, err)
		normalized, retryErr := n.Provider.GenerateBlueprint(ctx, retryPrompt)
		if retryErr != nil {
			return "", fmt.Errorf("AI normalization retry failed: %w", retryErr)
		}
		cleaned = ai.CleanAIOutput(normalized)
	}

	if err := ai.ValidateStructOutput(cleaned); err != nil {
		return "", fmt.Errorf("AI produced invalid output: %w", err)
	}

	return cleaned, nil
}

func (n *Normalizer) normalizeManual(input string) (string, error) {
	conv := New()

	tree, _, err := conv.Convert(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("manual normalization failed: %w", err)
	}

	normalized := conv.ConvertToString(tree)

	return normalized, nil
}

func (n *Normalizer) NormalizeToTree(ctx context.Context, input string) (*core.Tree, error) {
	normalized, err := n.Normalize(ctx, input)
	if err != nil {
		return nil, err
	}

	tree, err := n.Parser.ParseString(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("failed to parse normalized output: %w", err)
	}

	return tree, nil
}

func (n *Normalizer) DetectQuality(input string) int {
	score := 100

	if containsTreeSymbols(input) {
		score -= 30
	}
	if hasInconsistentIndentation(input) {
		score -= 20
	}
	if hasMixedSlashes(input) {
		score -= 15
	}
	if hasLineNumbers(input) {
		score -= 10
	}
	if len(input) > 5000 {
		score -= 10
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (n *Normalizer) SuggestMode(input string) NormalizationMode {
	quality := n.DetectQuality(input)

	if quality >= 70 {
		return ModeManual
	} else if quality >= 40 {
		return ModeAuto
	} else {
		return ModeAI
	}
}

func containsTreeSymbols(s string) bool {
	symbols := []string{"├", "└", "│", "─"}
	for _, sym := range symbols {
		if len(s) > 0 && contains(s, sym) {
			return true
		}
	}
	return false
}

func hasInconsistentIndentation(s string) bool {
	hasTab := false
	hasSpace := false

	lines := splitLines(s)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if line[0] == '\t' {
			hasTab = true
		} else if line[0] == ' ' {
			hasSpace = true
		}
		if hasTab && hasSpace {
			return true
		}
	}
	return false
}

func hasMixedSlashes(s string) bool {
	lines := splitLines(s)
	hasSlash := 0
	noSlash := 0

	for _, line := range lines {
		trimmed := trimLeft(line)
		if len(trimmed) == 0 {
			continue
		}

		if hasExtension(trimmed) {
			continue
		}

		if trimmed[len(trimmed)-1] == '/' {
			hasSlash++
		} else {
			noSlash++
		}
	}

	return hasSlash > 0 && noSlash > 0
}

func hasLineNumbers(s string) bool {
	lines := splitLines(s)
	for _, line := range lines {
		trimmed := trimLeft(line)
		if len(trimmed) > 0 && trimmed[0] >= '0' && trimmed[0] <= '9' {
			for i, ch := range trimmed {
				if ch == '.' || ch == ')' {
					if i > 0 && i < len(trimmed)-1 {
						return true
					}
				}
			}
		}
	}
	return false
}

func hasExtension(s string) bool {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i < len(s)-1
		}
		if s[i] == '/' || s[i] == '\\' {
			return false
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOfSubstring(s, substr) >= 0
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func trimLeft(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	return s[start:]
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else if ch != '\r' {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
