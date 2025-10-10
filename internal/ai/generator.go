package ai

import (
	"context"

	"github.com/alberdjuniawan/anstruct/internal/core"
)

type AIGenerator struct {
	Provider Provider
	Parser   core.Parser
}

func NewAIGenerator(p Provider, parser core.Parser) *AIGenerator {
	return &AIGenerator{Provider: p, Parser: parser}
}

// FromPrompt: natural language → blueprint tree
func (g *AIGenerator) FromPrompt(ctx context.Context, natural string) (*core.Tree, error) {
	text, err := g.Provider.GenerateBlueprint(ctx, natural)
	if err != nil {
		return nil, err
	}
	// langsung parse dari string, tanpa bikin file sementara
	return g.Parser.ParseString(ctx, text)
}
