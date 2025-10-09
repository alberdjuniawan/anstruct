package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGeminiProvider_GenerateBlueprint_Success(t *testing.T) {
	// fake server balikin blueprint sukses
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"blueprint": "app\n\tmain.go"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	p := NewGeminiProvider(ts.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := p.GenerateBlueprint(ctx, "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty blueprint")
	}
}

func TestGeminiProvider_GenerateBlueprint_ErrorFromProxy(t *testing.T) {
	// fake server balikin error status
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer ts.Close()

	p := NewGeminiProvider(ts.URL)
	ctx := context.Background()

	_, err := p.GenerateBlueprint(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGeminiProvider_GenerateBlueprint_EmptyBlueprint(t *testing.T) {
	// fake server balikin empty blueprint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"blueprint": ""}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	p := NewGeminiProvider(ts.URL)
	ctx := context.Background()

	_, err := p.GenerateBlueprint(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error for empty blueprint")
	}
}
