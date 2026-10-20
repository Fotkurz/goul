# Design

## Context

The project is greenfield (Go). The system serves as a self-hosted homelab alternative to goblin.tools' Magic ToDo, breaking down overwhelming tasks into actionable subtasks using a local Ollama LLM. The homelab host has constrained resources, requiring support for ultra-compact models (such as `qwen2.5:0.5b`) and a minimal deployment footprint. See `proposal.md` for motivation and `specs/` for behavioral requirements.

## Goals / Non-Goals

**Goals:**
- Single-binary deployment with zero external runtime frontend dependencies (embedded HTML/CSS/HTMX).
- Resilient structured outputs from local Ollama even when utilizing ultra-compact (~0.5B) models.
- Clean hierarchical task data modeling in SQLite supporting arbitrary subtask nesting.
- Fast, reactive UX using HTMX partial rendering with clear async loading states.

**Non-Goals:**
- Multi-user authentication, accounts, or role-based access control.
- Synchronization with third-party task platforms (Todoist, Google Tasks, Jira).
- Complex task scheduling, due dates, recurring reminders, or calendar integrations.

## Decisions

### 1. Web Architecture: Go standard library + HTMX + embedded assets
- **Decision**: Use Go 1.22+ `net/http` router with `html/template` and `embed.FS` to bundle HTMX, CSS, and HTML templates directly into the Go executable.
- **Rationale**: Keeps build and deployment as simple as a single static binary or tiny container image. HTMX provides snappy SPA-like UX for task list updates without a Node.js build pipeline or complex JavaScript frameworks.
- **Alternatives considered**:
  - *React/Svelte SPA*: Requires a Node/NPM build toolchain, asset management, and separate API routing.
  - *Server-rendered plain HTML without HTMX*: Full-page refreshes interrupt the user flow and reset scroll position when dissecting tasks.

### 2. Database Driver: Pure Go SQLite (`modernc.org/sqlite`)
- **Decision**: Use `modernc.org/sqlite` instead of CGO-based `mattn/go-sqlite3`.
- **Rationale**: CGO-free SQLite allows effortless cross-compilation across architectures (amd64, arm64 for Raspberry Pis and low-power homelab boxes) without needing target C toolchains or musl/glibc headaches.
- **Alternatives considered**:
  - *CGO `mattn/go-sqlite3`*: Standard choice in Go, but complicates cross-compilation and containerization.
  - *In-memory / JSON file*: Simple, but lacks cascading deletes, transactional safety, and query flexibility for tree hierarchies.

### 3. Data Schema: Adjacency List for Hierarchical Tasks
- **Decision**: Use an adjacency list schema with `parent_id` referencing `tasks(id) ON DELETE CASCADE`.
- **Schema**:
  ```sql
  CREATE TABLE IF NOT EXISTS tasks (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      parent_id INTEGER NULL REFERENCES tasks(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      completed BOOLEAN NOT NULL DEFAULT 0,
      position INTEGER NOT NULL DEFAULT 0,
      created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
  );
  CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);
  ```
- **Rationale**: An adjacency list is simple, directly maps to recursive task trees, and handles arbitrary nesting depth.

### 4. Ollama Integration: JSON Schema Structured Output
- **Decision**: Call Ollama's `/api/chat` using its `format` parameter with a JSON schema defining `{"steps": string[]}`, paired with a crisp, minimal system prompt.
- **Rationale**: Ultra-small models (~0.5B parameters) frequently deviate from markdown formatting or produce preamble conversational chatter. Constraining generation with Ollama's schema enforcement guarantees clean, parseable JSON arrays.
- **Prompt Strategy**:
  - *System*: "You are a task breakdown assistant. Provide actionable, concise steps to accomplish the given task. Return only the steps."
  - *User*: "Break down the following task into {N} steps: '{task}'" (where N is 3-4 for 'Nibble' and 6-8 for 'Feast').

### 5. Configuration Strategy: Environment Variables with Fallbacks
- **Decision**: Configure server port, database path, Ollama URL, and model via environment variables with sensible defaults:
  - `GOUL_PORT`: `:8080`
  - `GOUL_DB_PATH`: `./goul.db`
  - `OLLAMA_URL`: `http://localhost:11434`
  - `OLLAMA_MODEL`: `qwen2.5:0.5b`

## Risks / Trade-offs

- **[Risk] Ollama latency or cold-start on low-end hardware** → Mitigation: Use HTMX async request indicators (`hx-indicator`) showing a witty Halloween loading animation ("Ghoul is gnawing on this task...") with a sensible HTTP timeout (e.g. 60 seconds) so the UI remains responsive and does not freeze.
- **[Risk] Small model hallucinates unhelpful steps** → Mitigation: Keep prompts focused strictly on sequential actions, and allow users to delete, edit, or re-dissect any subtask easily in the UI.
- **[Risk] SQLite concurrency locks during simultaneous requests** → Mitigation: Enable WAL mode (`PRAGMA journal_mode=WAL;`) and configure a busy timeout on connection initialization.
