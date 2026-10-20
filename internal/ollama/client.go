package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Granularity string

const (
	GranularityNibble Granularity = "nibble"
	GranularityFeast  Granularity = "feast"
)

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewClient(baseURL, model string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) DissectTask(ctx context.Context, taskTitle string, granularity Granularity) ([]string, error) {
	trimmedTitle := strings.TrimSpace(taskTitle)
	if trimmedTitle == "" {
		return nil, fmt.Errorf("task title cannot be empty")
	}

	var userPrompt string
	switch strings.ToLower(string(granularity)) {
	case string(GranularityFeast):
		userPrompt = fmt.Sprintf("Break down the following task into 6 to 8 granular, bite-sized micro-steps: %q", trimmedTitle)
	default:
		userPrompt = fmt.Sprintf("Break down the following task into 3 to 4 concise, actionable milestones: %q", trimmedTitle)
	}

	reqBody := ChatRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are a task breakdown assistant. Provide actionable, sequential, plain text steps to accomplish the task. Return only the JSON object with the 'steps' list.",
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Stream: false,
		Format: DefaultTaskBreakdownSchema(),
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	apiURL := fmt.Sprintf("%s/api/chat", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("the ghoul could not reach Ollama at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response from ollama: %w", err)
	}

	content := strings.TrimSpace(chatResp.Message.Content)
	if content == "" {
		return nil, fmt.Errorf("ollama returned an empty message")
	}

	var result BreakdownResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse structured steps from ollama output: %w (raw output: %s)", err, content)
	}

	// Filter out any blank or duplicate whitespace strings
	var cleanSteps []string
	for _, step := range result.Steps {
		s := strings.TrimSpace(step)
		if s != "" {
			cleanSteps = append(cleanSteps, s)
		}
	}

	if len(cleanSteps) == 0 {
		return nil, fmt.Errorf("ollama generated no steps for task %q", trimmedTitle)
	}

	return cleanSteps, nil
}
