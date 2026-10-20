package ollama

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string         `json:"model"`
	Messages []ChatMessage  `json:"messages"`
	Stream   bool           `json:"stream"`
	Format   map[string]any `json:"format,omitempty"`
}

type ChatResponse struct {
	Model     string      `json:"model"`
	CreatedAt string      `json:"created_at"`
	Message   ChatMessage `json:"message"`
	Done      bool        `json:"done"`
}

type BreakdownResult struct {
	Steps []string `json:"steps"`
}

func DefaultTaskBreakdownSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"steps": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"steps"},
	}
}
