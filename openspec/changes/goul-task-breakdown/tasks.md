# Tasks

## 1. Project Scaffolding & Configuration

- [x] 1.1 Initialize the Go module (`go mod init goul`) and add `modernc.org/sqlite` dependency; verify compilation with `go build ./...`
- [x] 1.2 Implement configuration struct and loader reading `GOUL_PORT`, `GOUL_DB_PATH`, `OLLAMA_URL`, and `OLLAMA_MODEL` with fallback defaults; verify with unit tests in `internal/config/config_test.go`

## 2. Database & Storage Layer

- [x] 2.1 Implement SQLite database initialization with WAL mode, busy timeout, and schema creation for the hierarchical `tasks` table; verify connection in `internal/db/db_test.go`
- [x] 2.2 Implement Task repository operations (`CreateTask`, `GetTaskTree`, `ToggleTaskComplete`, `DeleteTask`); verify hierarchy retrieval and cascading delete in `internal/db/tasks_test.go`

## 3. Ollama Client & Breakdown Engine

- [x] 3.1 Define Ollama API request/response structs with JSON schema format constraints for structured array output; verify JSON serialization with unit tests
- [x] 3.2 Implement Ollama breakdown client supporting configurable granularity ("Nibble" vs "Feast") and prompt formatting; verify with mock HTTP server unit tests in `internal/ollama/client_test.go`
- [x] 3.3 Add error handling for unreachable Ollama daemon and malformed responses; verify error recovery in `internal/ollama/client_test.go`

## 4. HTTP Handlers & Web Interface

- [x] 4.1 Create embedded assets directory containing vendored HTMX script, CSS stylesheet with subtle Halloween/Ghoul dark styling, and HTML templates; verify assets embed correctly with `embed.FS`
- [x] 4.2 Implement HTTP handlers for root page (`GET /`), task creation (`POST /tasks`), status toggle (`PATCH /tasks/{id}/toggle`), and deletion (`DELETE /tasks/{id}`); verify endpoint responses with `net/http/httptest`
- [x] 4.3 Implement dissection handler (`POST /tasks/{id}/dissect`) calling the Ollama client, saving resulting subtasks to SQLite, and returning the updated task tree partial; verify with handler tests
- [x] 4.4 Implement UI loading state indicators ("Ghoul is digging...") using `hx-indicator` and error toast/banner for Ollama connectivity issues; verify UI template rendering

## 5. End-to-End Verification & Documentation

- [x] 5.1 Create application entrypoint in `cmd/goul/main.go` wiring config, database, Ollama client, and HTTP server; verify binary builds successfully with `go build -o goul ./cmd/goul`
- [x] 5.2 Write `README.md` documenting quickstart, configuration environment variables, homelab deployment, and Ollama model setup; verify all documented commands execute as expected
