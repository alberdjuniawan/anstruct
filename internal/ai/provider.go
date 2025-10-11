package ai

import "context"

type Provider interface {
	GenerateBlueprint(ctx context.Context, prompt string) (string, error)
}
