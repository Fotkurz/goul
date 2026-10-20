# Spec Delta

## Purpose

Enables breaking down complex or intimidating tasks into ordered, bite-sized subtasks using a local Ollama LLM with configurable granularity.

## ADDED Requirements

### Requirement: Task Dissection via Local LLM
The system SHALL request a breakdown of a given task from a local Ollama instance and insert the resulting steps as ordered subtasks under the targeted task.

#### Scenario: Successful task breakdown
- **WHEN** the user triggers dissection on a task
- **THEN** the system contacts Ollama with the task title, parses the structured response steps, saves them as child tasks, and updates the task tree view

### Requirement: Granularity Control
The system SHALL allow the user to select the granularity level ("hunger level") before requesting a dissection to control the depth and step count.

#### Scenario: Nibble granularity
- **WHEN** a dissection is requested with low granularity ("Nibble")
- **THEN** the prompt requests 3 to 4 high-level actionable milestones

#### Scenario: Feast granularity
- **WHEN** a dissection is requested with high granularity ("Feast")
- **THEN** the prompt requests 6 to 8 granular micro-steps designed to reduce executive overwhelm

### Requirement: Recursive Dissection
The system SHALL allow subtasks to be dissected further into sub-subtasks.

#### Scenario: Dissecting an existing subtask
- **WHEN** the user triggers dissection on an existing subtask
- **THEN** the system generates child tasks parented to that subtask, creating a nested branch in the hierarchy

### Requirement: Error Handling and Availability
The system SHALL handle Ollama communication timeouts, connection refusals, and malformed responses gracefully without corrupting task state.

#### Scenario: Ollama service unavailable
- **WHEN** a dissection is requested but the Ollama endpoint is unreachable
- **THEN** the system displays a clear, witty error notification explaining Ollama could not be reached, preserving all existing tasks intact
