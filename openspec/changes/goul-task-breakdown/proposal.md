# Proposal

## Why

Users who struggle with executive dysfunction, task paralysis, or overwhelming to-do lists need a simple, single-purpose tool to dissect high-level tasks into manageable subtasks. While tools like goblin.tools offer this in the cloud, homelab users need a lightweight, self-hosted, privacy-first alternative that operates entirely locally on Ollama without external dependencies or heavy hardware requirements.

## What Changes

- Introduce a lightweight self-hosted Go web application with embedded HTMX and CSS.
- Provide a responsive, witty, ghoul-themed interface for adding tasks, toggling completion, deleting tasks, and breaking tasks down.
- Implement integration with local Ollama instances (defaulting to lightweight models such as `qwen2.5:0.5b`) with structured JSON schema responses to dissect tasks into sequential subtasks.
- Provide granularity control ("Ghoul Hunger": e.g., Nibble vs. Feast) to determine the depth/step-count of the breakdown.
- Support recursive dissection: subtasks can themselves be broken down into nested sub-subtasks.
- Persist hierarchical task state in a local SQLite database (`goul.db`).

## Capabilities

### New Capabilities
- `task-management`: Managing hierarchical tasks and subtasks (creation, completion toggle, deletion, reordering, retrieval) backed by SQLite.
- `task-breakdown`: Deconstructing a task into actionable steps using a local Ollama LLM with configurable granularity ("ghoul hunger") and structured JSON outputs.

### Modified Capabilities
None (this is the initial capability set for the project).

## Impact

- Greenfield project setup: Go module initialization (`go.mod`), HTTP routing, SQLite schema setup, embedded templates/assets.
- Requires local Ollama running on the homelab network or local host.
- No breaking changes.
