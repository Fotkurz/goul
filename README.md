# 🪦 GOUL [o.o]

> **"Dissect your overwhelming tasks into digestible bones."**

**Goul** is a lightweight, self-hosted task breakdown tool inspired by [goblin.tools](https://goblin.tools)' Magic ToDo. Built for homelabs and privacy enthusiasts, Goul runs completely on your local network and connects to local LLMs via [Ollama](https://ollama.ai).

It is designed to be bone-chillingly simple, resource-efficient, and capable of operating even on low-end hardware using ultra-compact local models (like `qwen2.5:0.5b` or `llama3.2:1b`).

---

## ✨ Features

- 🦴 **AI Task Dissection**: Click "Dissect" to break any high-level chore or daunting project into concrete, actionable steps.
- 🍽️ **Ghoul Appetite (Granularity)**:
  - **Nibble**: Generates 3–4 high-level milestones.
  - **Feast**: Generates 6–8 granular micro-steps designed to overcome executive dysfunction.
- 🌳 **Recursive Dissection**: Any subtask can itself be dissected into deeper sub-subtasks.
- 🪦 **SQLite Persistence**: Hierarchical adjacency tree storage in a single SQLite file with cascading deletions ("Bury").
- ⚡ **Zero-Dependency Single Binary**: Pure Go, CGO-free, with HTML templates, CSS, and vendored HTMX embedded directly into the executable via `embed.FS`.
- 🎃 **Subtle & Witty Halloween Theme**: High contrast dark crypt UI with responsive design.

---

## 🚀 Quickstart

### Prerequisites

1. **Go 1.22+** installed on your system.
2. **Ollama** running locally or in your homelab:
   ```bash
   # Pull a lightweight model (works great on low-power CPUs/GPUs)
   ollama pull qwen2.5:0.5b
   ```

### Running Locally

```bash
# Build the single standalone binary
go build -o goul ./cmd/goul

# Run the server
./goul
```

Open your browser at **[http://localhost:8080](http://localhost:8080)**.

---

## ⚙️ Configuration

Goul is configured via standard environment variables:

| Variable | Default | Description |
|---|---|---|
| `GOUL_PORT` / `PORT` | `:8080` | Port the HTTP web server binds to |
| `GOUL_DB_PATH` | `./goul.db` | File path to the SQLite database |
| `OLLAMA_URL` | `http://localhost:11434` | URL of your local Ollama instance |
| `OLLAMA_MODEL` | `qwen2.5:0.5b` | Model name to use for task breakdown |

### Example with Custom Model & Host

```bash
OLLAMA_URL="http://192.168.1.100:11434" \
OLLAMA_MODEL="llama3.2:3b" \
GOUL_PORT=":9000" \
./goul
```

---

## 🧪 Testing

Run the full suite of unit and integration tests across all packages:

```bash
go test -v ./...
```

---

## 🏠 Homelab Deployment

### Systemd Service

Create `/etc/systemd/system/goul.service`:

```ini
[Unit]
Description=Goul Task Breakdown Service
After=network.target ollama.service

[Service]
Type=simple
User=homelab
WorkingDirectory=/opt/goul
ExecStart=/opt/goul/goul
Restart=on-failure
Environment=GOUL_PORT=:8080
Environment=OLLAMA_URL=http://127.0.0.1:11434
Environment=OLLAMA_MODEL=qwen2.5:0.5b
Environment=GOUL_DB_PATH=/opt/goul/data/goul.db

[Install]
WantedBy=multi-user.target
```

Reload and enable:
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now goul
```

---

## 📜 License

MIT
