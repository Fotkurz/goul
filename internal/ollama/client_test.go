package ollama

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDissectTask_GranularityAndSuccess(t *testing.T) {
	var capturedReq ChatRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			http.NotFound(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed reading request: %v", err)
		}
		if err := json.Unmarshal(body, &capturedReq); err != nil {
			t.Fatalf("failed unmarshaling request: %v", err)
		}

		resp := ChatResponse{
			Model: "test-model",
			Message: ChatMessage{
				Role:    "assistant",
				Content: `{"steps": ["Clear floor", "Bag trash", "Sweep dust"]}`,
			},
			Done: true,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model", 2*time.Second)
	ctx := context.Background()

	// Test 1: Nibble granularity
	steps, err := client.DissectTask(ctx, "Clean the garage", GranularityNibble)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}
	if !strings.Contains(capturedReq.Messages[1].Content, "concise, actionable milestones") {
		t.Errorf("expected Nibble prompt, got: %s", capturedReq.Messages[1].Content)
	}

	// Test 2: Feast granularity
	_, err = client.DissectTask(ctx, "Clean the garage", GranularityFeast)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedReq.Messages[1].Content, "granular, bite-sized micro-steps") {
		t.Errorf("expected Feast prompt, got: %s", capturedReq.Messages[1].Content)
	}
}

func TestDissectTask_Errors(t *testing.T) {
	ctx := context.Background()

	// Test 1: Unreachable Ollama daemon
	unreachableClient := NewClient("http://127.0.0.1:59999", "model", 100*time.Millisecond)
	_, err := unreachableClient.DissectTask(ctx, "Test task", GranularityNibble)
	if err == nil || !strings.Contains(err.Error(), "the ghoul could not reach Ollama") {
		t.Errorf("expected unreachable error, got: %v", err)
	}

	// Test 2: HTTP 500 from Ollama
	server500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server500.Close()

	client500 := NewClient(server500.URL, "model", time.Second)
	_, err = client500.DissectTask(ctx, "Test task", GranularityNibble)
	if err == nil || !strings.Contains(err.Error(), "ollama returned HTTP 500") {
		t.Errorf("expected HTTP 500 error, got: %v", err)
	}

	// Test 3: Malformed JSON from model
	serverMalformed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Model: "test-model",
			Message: ChatMessage{
				Role:    "assistant",
				Content: "I cannot do that.",
			},
			Done: true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverMalformed.Close()

	clientMalformed := NewClient(serverMalformed.URL, "model", time.Second)
	_, err = clientMalformed.DissectTask(ctx, "Test task", GranularityNibble)
	if err == nil || !strings.Contains(err.Error(), "failed to parse structured steps") {
		t.Errorf("expected malformed steps error, got: %v", err)
	}
}
