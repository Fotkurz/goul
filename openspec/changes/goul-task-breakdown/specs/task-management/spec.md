# Spec Delta

## Purpose

Enables users to create, view, complete, and delete hierarchical to-do tasks and subtasks persisted across sessions.

## ADDED Requirements

### Requirement: Task Creation
The system SHALL allow users to create top-level tasks by providing a non-empty text title.

#### Scenario: Successful task creation
- **WHEN** the user submits a valid task title in the task input form
- **THEN** the system persists the task and renders it at the top of the task list

#### Scenario: Empty task title submission
- **WHEN** the user submits an empty or whitespace-only task title
- **THEN** the system rejects the creation without altering stored tasks or displaying a blank item

### Requirement: Task Completion Status
The system SHALL allow users to toggle the completed status of any task or subtask.

#### Scenario: Marking task as completed
- **WHEN** the user toggles an uncompleted task checkbox
- **THEN** the system marks the task as completed, reflects the status visually, and persists the update

#### Scenario: Unmarking a completed task
- **WHEN** the user unchecks a completed task
- **THEN** the system restores the task to uncompleted status and persists the change

### Requirement: Task Deletion
The system SHALL allow users to permanently delete ("bury") any task, cascading to all of its nested subtasks.

#### Scenario: Deleting a parent task with subtasks
- **WHEN** the user triggers deletion on a parent task
- **THEN** the system removes the parent task and all associated descendants from storage and removes them from the view

### Requirement: Hierarchical Task Nesting
The system SHALL support nesting subtasks under parent tasks at arbitrary depth.

#### Scenario: Displaying nested subtasks
- **WHEN** tasks have child subtasks associated by parent identifier
- **THEN** the system renders child subtasks indented directly beneath their respective parent task in sorted order
