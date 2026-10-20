package ollama

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestChatRequestSerialization(t *testing.T) {
	req := ChatRequest{
		Model: "qwen2.5:0.5b",
		Messages: []ChatMessage{
			{Role: "system", Content: "Break down the task."},
			{Role: "user", Content: "Clean my desk"},
		},
		Stream: false,
		Format: DefaultTaskBreakdownSchema(),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal ChatRequest: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"model":"qwen2.5:0.5b"`) {
		t.Errorf("expected model in json, got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"steps"`) {
		t.Errorf("expected format steps in json, got %s", jsonStr)
	}
}

func TestChatResponseDeserialization(t *testing.T) {
	rawResponse := `{
		"model": "qwen2.5:0.5b",
		"created_at": "2026-10-20T00:00:00Z",
		"message": {
			"role": "assistant",
			"content": "{\"steps\":[\"Remove empty mugs\",\"Wipe surface\",\"Organize cables\"]}"
		},
		"done": true
	}`

	var res ChatResponse
	if err := json.Unmarshal([]byte(rawResponse), &res); err != nil {
		t.Fatalf("failed to unmarshal ChatResponse: %v", err)
	}

	if res.Message.Role != "assistant" {
		t.Errorf("expected assistant role, got %s", res.Message.Role)
	}

	var breakdown BreakdownResult
	if err := json.Unmarshal([]byte(res.Message.Content), &breakdown); err != nil {
		t.Fatalf("failed to unmarshal inner steps JSON: %v", err)
	}

	if len(breakdown.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(breakdown.Steps))
	}
	if breakdown.Steps[0] != "Remove empty mugs" {
		t.Errorf("expected 'Remove empty mugs', got %s", breakdown.Steps[0])
	}
}
