package ai

import (
	"bytes"
	"context"
	"encoding/json"
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
func NewGeminiProvider(endpoint string) *GeminiProvider {
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

// GenerateBlueprint mengirim prompt alami ke proxy → menerima blueprint .struct
func (g *GeminiProvider) GenerateBlueprint(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(requestPayload{Prompt: prompt})

	req, err := http.NewRequestWithContext(ctx, "POST", g.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("proxy error: %s (%s)", resp.Status, string(b))
	}

	var result responsePayload
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("proxy returned error: %s", result.Error)
	}
	if result.Blueprint == "" {
		return "", fmt.Errorf("empty blueprint returned")
	}
	return result.Blueprint, nil
}
