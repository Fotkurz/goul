package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"goul/internal/db"
	"goul/internal/ollama"
)

func setupTestServer(t *testing.T, ollamaServerURL string) (*Server, *db.DB) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "server_test.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() {
		database.Close()
	})

	ollamaClient := ollama.NewClient(ollamaServerURL, "qwen2.5:0.5b", 2*time.Second)
	srv, err := NewServer(database, ollamaClient, "qwen2.5:0.5b")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	return srv, database
}

func TestServer_RootAndCRUD(t *testing.T) {
	srv, database := setupTestServer(t, "http://localhost:11434")

	// 1. GET /
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for GET /, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "GOUL") {
		t.Errorf("expected GOUL in response body")
	}

	// 2. POST /tasks (empty title error)
	form := url.Values{"title": {"   "}}
	req = httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for empty title, got %d", w.Code)
	}

	// 3. POST /tasks (valid task)
	form = url.Values{"title": {"Carve pumpkin"}}
	req = httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for valid task, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Carve pumpkin") {
		t.Errorf("expected Carve pumpkin in response body")
	}

	// Query created task
	tree, err := database.GetTaskTree(context.Background())
	if err != nil || len(tree) != 1 {
		t.Fatalf("expected 1 task in database, got %d", len(tree))
	}
	taskID := tree[0].ID

	// 4. PATCH /tasks/{id}/toggle
	req = httptest.NewRequest(http.MethodPatch, "/tasks/"+strconv.FormatInt(taskID, 10)+"/toggle", nil)
	w = httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for toggle, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "completed") {
		t.Errorf("expected task to have completed class")
	}

	// 5. DELETE /tasks/{id}
	req = httptest.NewRequest(http.MethodDelete, "/tasks/"+strconv.FormatInt(taskID, 10), nil)
	w = httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for delete, got %d", w.Code)
	}

	treeAfter, err := database.GetTaskTree(context.Background())
	if err != nil || len(treeAfter) != 0 {
		t.Fatalf("expected 0 tasks in database after delete, got %d", len(treeAfter))
	}
}

func TestServer_DissectHandler(t *testing.T) {
	// Mock Ollama server
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "ErrorTrigger") {
			http.Error(w, "model failed", http.StatusInternalServerError)
			return
		}

		resp := ollama.ChatResponse{
			Model: "qwen2.5:0.5b",
			Message: ollama.ChatMessage{
				Role:    "assistant",
				Content: `{"steps": ["Buy spooky seeds", "Dig a plot", "Water with graveyard mist"]}`,
			},
			Done: true,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockOllama.Close()

	srv, database := setupTestServer(t, mockOllama.URL)

	// Create a task to dissect
	task, err := database.CreateTask(context.Background(), nil, "Plant nightshade")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Dissect with Feast granularity
	form := url.Values{"hunger": {"feast"}}
	req := httptest.NewRequest(http.MethodPost, "/tasks/"+strconv.FormatInt(task.ID, 10)+"/dissect", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for dissect, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "Buy spooky seeds") || !strings.Contains(body, "Water with graveyard mist") {
		t.Errorf("expected generated subtasks in rendered tree, got: %s", body)
	}

	// Verify in database
	tree, err := database.GetTaskTree(context.Background())
	if err != nil || len(tree) != 1 || len(tree[0].Children) != 3 {
		t.Fatalf("expected 1 root task with 3 children, got %+v", tree)
	}

	// Test error case
	errTask, err := database.CreateTask(context.Background(), nil, "ErrorTrigger Task")
	if err != nil {
		t.Fatalf("failed to create errTask: %v", err)
	}

	req = httptest.NewRequest(http.MethodPost, "/tasks/"+strconv.FormatInt(errTask.ID, 10)+"/dissect", nil)
	w = httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "error-banner") {
		t.Errorf("expected error-banner in response on Ollama error, got %s", w.Body.String())
	}
}
