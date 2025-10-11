package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GeminiProvider struct {
	Endpoint string
	Client   *http.Client
}

func NewGeminiProvider(endpoint string) *GeminiProvider {
	if endpoint == "" {
		endpoint = "https://anstruct-ai-proxy.anstruct.workers.dev/generate"
	}
	return &GeminiProvider{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}

type requestPayload struct {
	Prompt string `json:"prompt"`
}

type responsePayload struct {
	Blueprint string `json:"blueprint"`
	Error     string `json:"error,omitempty"`
}

var ErrEmptyBlueprint = errors.New("empty blueprint returned")

func (g *GeminiProvider) GenerateBlueprint(ctx context.Context, userPrompt string) (string, error) {
	if userPrompt == "" {
		return "", errors.New("prompt cannot be empty")
	}

	fullPrompt := BuildFullPrompt(userPrompt)

	body, err := json.Marshal(requestPayload{Prompt: fullPrompt})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("proxy error: %s (%s)", resp.Status, string(data))
	}

	var result responsePayload
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("proxy returned error: %s", result.Error)
	}
	if result.Blueprint == "" {
		return "", ErrEmptyBlueprint
	}
	return result.Blueprint, nil
}
