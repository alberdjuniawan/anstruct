package ai

import (
	"context"
	"fmt"

	"github.com/alberdjuniawan/anstruct/internal/core"
)

type AIGenerator struct {
	Provider Provider
	Parser   core.Parser
}

func NewAIGenerator(p Provider, parser core.Parser) *AIGenerator {
	return &AIGenerator{Provider: p, Parser: parser}
}

func (g *AIGenerator) FromPrompt(ctx context.Context, natural string, retries int) (*core.Tree, string, error) {
	if retries < 1 {
		retries = 1
	}

	var lastErr error
	currentPrompt := natural

	for attempt := 1; attempt <= retries; attempt++ {
		text, err := g.Provider.GenerateBlueprint(ctx, currentPrompt)
		if err != nil {
			lastErr = fmt.Errorf("AI provider error (attempt %d/%d): %w", attempt, retries, err)
			continue
		}

		cleanedText := CleanAIOutput(text)

		cleanedText = DetectAndWrapSingleRoot(cleanedText, "project")

		if err := ValidateStructOutput(cleanedText); err != nil {
			lastErr = fmt.Errorf("validation error (attempt %d/%d): %w", attempt, retries, err)

			if attempt < retries {
				currentPrompt = RetryPrompt(natural, err)
				fmt.Printf("⚠️ Attempt %d failed, retrying with corrections...\n", attempt)
				continue
			}
			return nil, cleanedText, lastErr
		}

		tree, err := g.Parser.ParseString(ctx, cleanedText)
		if err != nil {
			lastErr = fmt.Errorf("parse error (attempt %d/%d): %w", attempt, retries, err)
			if attempt < retries {
				currentPrompt = RetryPrompt(natural, err)
				fmt.Printf("⚠️ Attempt %d failed, retrying...\n", attempt)
				continue
			}
			return nil, cleanedText, lastErr
		}

		if attempt > 1 {
			fmt.Printf("✅ Succeeded on attempt %d\n", attempt)
		}
		return tree, cleanedText, nil
	}

	return nil, "", fmt.Errorf("all %d attempts failed: %w", retries, lastErr)
}

func (g *AIGenerator) FromPromptSimple(ctx context.Context, natural string) (*core.Tree, error) {
	tree, _, err := g.FromPrompt(ctx, natural, 2)
	return tree, err
}
