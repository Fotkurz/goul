package web

import (
	"bytes"
	"goul/internal/db"
	"testing"
)

func TestParseTemplates(t *testing.T) {
	tmpl, err := ParseTemplates()
	if err != nil {
		t.Fatalf("failed to parse embedded templates: %v", err)
	}

	// Verify executing index.html
	var buf bytes.Buffer
	data := struct {
		Model string
		Tasks []*db.Task
	}{
		Model: "qwen2.5:0.5b",
		Tasks: []*db.Task{
			{
				ID:        1,
				Title:     "Root Task",
				Completed: false,
				Children: []*db.Task{
					{
						ID:        2,
						Title:     "Subtask",
						Completed: true,
					},
				},
			},
		},
	}

	if err := tmpl.ExecuteTemplate(&buf, "index.html", data); err != nil {
		t.Fatalf("failed to execute index.html template: %v", err)
	}

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("Root Task")) {
		t.Errorf("expected Root Task in output")
	}
	if !bytes.Contains([]byte(out), []byte("Subtask")) {
		t.Errorf("expected Subtask in output")
	}
}

func TestStaticFilesEmbedded(t *testing.T) {
	cssData, err := StaticFS.ReadFile("static/style.css")
	if err != nil {
		t.Fatalf("failed to read static/style.css from embed: %v", err)
	}
	if len(cssData) == 0 {
		t.Errorf("style.css is empty")
	}

	jsData, err := StaticFS.ReadFile("static/htmx.min.js")
	if err != nil {
		t.Fatalf("failed to read static/htmx.min.js from embed: %v", err)
	}
	if len(jsData) == 0 {
		t.Errorf("htmx.min.js is empty")
	}
}
