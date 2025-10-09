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

// GeminiProvider memanggil proxy endpoint (misalnya Cloudflare Worker)
// yang menyimpan API key Gemini secara aman di server.
type GeminiProvider struct {
	Endpoint string       // URL proxy, contoh: https://anstruct-ai-proxy.workers.dev/generate
	Client   *http.Client // custom HTTP client
}

// NewGeminiProvider membuat provider baru dengan endpoint proxy.
// Jika endpoint kosong, gunakan default.
func NewGeminiProvider(endpoint string) *GeminiProvider {
	if endpoint == "" {
		endpoint = "https://anstruct-ai-proxy.anstruct.workers.dev/generate"
	}
	return &GeminiProvider{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// requestPayload adalah format request ke proxy
type requestPayload struct {
	Prompt string `json:"prompt"`
}

// responsePayload adalah format response dari proxy
type responsePayload struct {
	Blueprint string `json:"blueprint"`
	Error     string `json:"error,omitempty"`
}

// ErrEmptyBlueprint error khusus jika hasil kosong
var ErrEmptyBlueprint = errors.New("empty blueprint returned")

// GenerateBlueprint mengirim prompt alami ke proxy → menerima blueprint .struct
func (g *GeminiProvider) GenerateBlueprint(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", errors.New("prompt cannot be empty")
	}

	body, err := json.Marshal(requestPayload{Prompt: prompt})
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

	// baca body sekali
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
