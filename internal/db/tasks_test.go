package db

import (
	"context"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	t.Cleanup(func() {
		database.Close()
	})
	return database
}

func TestTasksCRUD(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	// 1. Test validation on empty title
	_, err := database.CreateTask(ctx, nil, "   ")
	if err == nil {
		t.Error("expected error for empty task title, got nil")
	}

	// 2. Test create root tasks
	t1, err := database.CreateTask(ctx, nil, "Root Task 1")
	if err != nil {
		t.Fatalf("failed to create root task 1: %v", err)
	}
	if t1.Title != "Root Task 1" || t1.Completed || t1.ParentID != nil {
		t.Errorf("unexpected task fields: %+v", t1)
	}

	t2, err := database.CreateTask(ctx, nil, "Root Task 2")
	if err != nil {
		t.Fatalf("failed to create root task 2: %v", err)
	}

	// 3. Test create subtasks
	sub1, err := database.CreateTask(ctx, &t1.ID, "Subtask 1.1")
	if err != nil {
		t.Fatalf("failed to create subtask: %v", err)
	}
	if sub1.ParentID == nil || *sub1.ParentID != t1.ID {
		t.Errorf("expected parent ID %d, got %+v", t1.ID, sub1.ParentID)
	}

	subSub1, err := database.CreateTask(ctx, &sub1.ID, "Sub-subtask 1.1.1")
	if err != nil {
		t.Fatalf("failed to create sub-subtask: %v", err)
	}

	// 4. Test GetTaskTree hierarchy
	tree, err := database.GetTaskTree(ctx)
	if err != nil {
		t.Fatalf("failed to get task tree: %v", err)
	}

	if len(tree) != 2 {
		t.Fatalf("expected 2 root tasks, got %d", len(tree))
	}

	// Find Root Task 1 in tree
	var rootFound *Task
	for _, root := range tree {
		if root.ID == t1.ID {
			rootFound = root
			break
		}
	}
	if rootFound == nil {
		t.Fatalf("Root Task 1 not found in tree")
	}
	if len(rootFound.Children) != 1 {
		t.Fatalf("expected 1 child for root 1, got %d", len(rootFound.Children))
	}
	if len(rootFound.Children[0].Children) != 1 {
		t.Fatalf("expected 1 child for subtask 1.1, got %d", len(rootFound.Children[0].Children))
	}
	if rootFound.Children[0].Children[0].ID != subSub1.ID {
		t.Errorf("expected sub-subtask id %d, got %d", subSub1.ID, rootFound.Children[0].Children[0].ID)
	}

	// 5. Test toggle completion
	completed, err := database.ToggleTaskComplete(ctx, t1.ID)
	if err != nil {
		t.Fatalf("failed to toggle completion: %v", err)
	}
	if !completed {
		t.Errorf("expected task to be completed, got %v", completed)
	}

	completed, err = database.ToggleTaskComplete(ctx, t1.ID)
	if err != nil {
		t.Fatalf("failed to toggle completion again: %v", err)
	}
	if completed {
		t.Errorf("expected task to be uncompleted, got %v", completed)
	}

	// 6. Test cascading delete
	err = database.DeleteTask(ctx, t1.ID)
	if err != nil {
		t.Fatalf("failed to delete root task 1: %v", err)
	}

	// Verify t1, sub1, and subSub1 are deleted
	_, err = database.GetTaskByID(ctx, t1.ID)
	if err == nil {
		t.Errorf("expected t1 to be deleted")
	}
	_, err = database.GetTaskByID(ctx, sub1.ID)
	if err == nil {
		t.Errorf("expected sub1 to be deleted via cascade")
	}
	_, err = database.GetTaskByID(ctx, subSub1.ID)
	if err == nil {
		t.Errorf("expected subSub1 to be deleted via cascade")
	}

	// t2 should still exist
	t2Check, err := database.GetTaskByID(ctx, t2.ID)
	if err != nil || t2Check == nil {
		t.Errorf("expected t2 to still exist")
	}
}
