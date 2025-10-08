package ai

import "context"

// Provider adalah kontrak umum untuk semua AI provider.
// Bisa diimplementasikan oleh Gemini, OpenAI, atau bahkan mock untuk testing.
type Provider interface {
	GenerateBlueprint(ctx context.Context, prompt string) (string, error)
}
