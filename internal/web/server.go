package web

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"goul/internal/db"
	"goul/internal/ollama"
)

type Server struct {
	db     *db.DB
	ollama *ollama.Client
	model  string
	tmpl   *template.Template
	mux    *http.ServeMux
}

func NewServer(database *db.DB, ollamaClient *ollama.Client, model string) (*Server, error) {
	tmpl, err := ParseTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	s := &Server{
		db:     database,
		ollama: ollamaClient,
		model:  model,
		tmpl:   tmpl,
		mux:    http.NewServeMux(),
	}

	s.registerRoutes()
	return s, nil
}

func (s *Server) Routes() http.Handler {
	return s.mux
}

func (s *Server) registerRoutes() {
	staticSubFS, err := fs.Sub(StaticFS, "static")
	if err != nil {
		log.Fatalf("failed to create static sub-fs: %v", err)
	}
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS))))

	s.mux.HandleFunc("GET /", s.handleRoot)
	s.mux.HandleFunc("POST /tasks", s.handleCreateTask)
	s.mux.HandleFunc("PATCH /tasks/{id}/toggle", s.handleToggleTask)
	s.mux.HandleFunc("DELETE /tasks/{id}", s.handleDeleteTask)
	s.mux.HandleFunc("POST /tasks/{id}/dissect", s.handleDissectTask)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tasks, err := s.db.GetTaskTree(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get tasks: %v", err), http.StatusInternalServerError)
		return
	}

	data := struct {
		Model string
		Tasks []*db.Task
	}{
		Model: s.model,
		Tasks: tasks,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = s.renderError(w, "Task title cannot be empty")
		return
	}

	var parentID *int64
	if pStr := r.FormValue("parent_id"); pStr != "" {
		if pid, err := strconv.ParseInt(pStr, 10, 64); err == nil {
			parentID = &pid
		}
	}

	if _, err := s.db.CreateTask(r.Context(), parentID, title); err != nil {
		http.Error(w, fmt.Sprintf("failed to create task: %v", err), http.StatusInternalServerError)
		return
	}

	s.renderTaskTree(w, r.Context())
}

func (s *Server) handleToggleTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	if _, err := s.db.ToggleTaskComplete(r.Context(), id); err != nil {
		http.Error(w, fmt.Sprintf("failed to toggle task: %v", err), http.StatusInternalServerError)
		return
	}

	task, err := s.db.GetTaskByID(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to retrieve task: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "task_item.html", task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	if err := s.db.DeleteTask(r.Context(), id); err != nil {
		http.Error(w, fmt.Sprintf("failed to delete task: %v", err), http.StatusInternalServerError)
		return
	}

	// Returning empty response causes outerHTML swap to remove the element from the DOM
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDissectTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	hunger := r.FormValue("hunger")
	if hunger == "" {
		hunger = string(ollama.GranularityNibble)
	}

	task, err := s.db.GetTaskByID(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("task not found: %v", err), http.StatusNotFound)
		return
	}

	steps, err := s.ollama.DissectTask(r.Context(), task.Title, ollama.Granularity(hunger))
	if err != nil {
		// Return error toast/banner
		w.Header().Set("HX-Reswap", "innerHTML")
		w.Header().Set("HX-Retarget", "#error-container")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = s.renderError(w, err.Error())
		return
	}

	// Persist subtasks in SQLite
	for _, step := range steps {
		if _, err := s.db.CreateTask(r.Context(), &task.ID, step); err != nil {
			log.Printf("failed to save subtask %q: %v", step, err)
		}
	}

	s.renderTaskTree(w, r.Context())
}

func (s *Server) renderTaskTree(w http.ResponseWriter, ctx context.Context) {
	tasks, err := s.db.GetTaskTree(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get task tree: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "task_tree.html", tasks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderError(w http.ResponseWriter, message string) error {
	tmpl := `<div class="error-banner">
  <span>⚠️ %s</span>
  <button type="button" class="btn btn-ghost" onclick="this.parentElement.remove()">Dismiss</button>
</div>`
	_, err := fmt.Fprintf(w, tmpl, template.HTMLEscapeString(message))
	return err
}
