package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Task struct {
	ID        int64     `json:"id"`
	ParentID  *int64    `json:"parent_id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	Children  []*Task   `json:"children,omitempty"`
}

func (db *DB) CreateTask(ctx context.Context, parentID *int64, title string) (*Task, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return nil, fmt.Errorf("task title cannot be empty")
	}

	var maxPos sql.NullInt64
	var err error
	if parentID == nil {
		err = db.QueryRowContext(ctx, "SELECT MAX(position) FROM tasks WHERE parent_id IS NULL").Scan(&maxPos)
	} else {
		err = db.QueryRowContext(ctx, "SELECT MAX(position) FROM tasks WHERE parent_id = ?", *parentID).Scan(&maxPos)
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query max position: %w", err)
	}

	pos := 0
	if maxPos.Valid {
		pos = int(maxPos.Int64) + 1
	}

	var res sql.Result
	if parentID == nil {
		res, err = db.ExecContext(ctx, "INSERT INTO tasks (parent_id, title, position) VALUES (NULL, ?, ?)", trimmed, pos)
	} else {
		res, err = db.ExecContext(ctx, "INSERT INTO tasks (parent_id, title, position) VALUES (?, ?, ?)", *parentID, trimmed, pos)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return db.GetTaskByID(ctx, id)
}

func (db *DB) GetTaskByID(ctx context.Context, id int64) (*Task, error) {
	row := db.QueryRowContext(ctx, "SELECT id, parent_id, title, completed, position, created_at FROM tasks WHERE id = ?", id)
	var t Task
	var parentID sql.NullInt64
	var completedInt int
	var createdAt string

	if err := row.Scan(&t.ID, &parentID, &t.Title, &completedInt, &t.Position, &createdAt); err != nil {
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	if parentID.Valid {
		val := parentID.Int64
		t.ParentID = &val
	}
	t.Completed = completedInt != 0
	parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt)
	if err == nil {
		t.CreatedAt = parsedTime
	} else {
		t.CreatedAt = time.Now()
	}

	return &t, nil
}

func (db *DB) GetTaskTree(ctx context.Context) ([]*Task, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, parent_id, title, completed, position, created_at FROM tasks ORDER BY position ASC, id ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	taskMap := make(map[int64]*Task)
	var roots []*Task

	for rows.Next() {
		var t Task
		var parentID sql.NullInt64
		var completedInt int
		var createdAt string

		if err := rows.Scan(&t.ID, &parentID, &t.Title, &completedInt, &t.Position, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}

		if parentID.Valid {
			val := parentID.Int64
			t.ParentID = &val
		}
		t.Completed = completedInt != 0
		parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt)
		if err == nil {
			t.CreatedAt = parsedTime
		} else {
			t.CreatedAt = time.Now()
		}
		t.Children = make([]*Task, 0)

		taskMap[t.ID] = &t
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading task rows: %w", err)
	}

	// Build tree hierarchy
	for _, task := range taskMap {
		if task.ParentID == nil {
			roots = append(roots, task)
		} else {
			parent, exists := taskMap[*task.ParentID]
			if exists {
				parent.Children = append(parent.Children, task)
			} else {
				// Parent orphaned, treat as root
				roots = append(roots, task)
			}
		}
	}

	return roots, nil
}

func (db *DB) ToggleTaskComplete(ctx context.Context, id int64) (bool, error) {
	_, err := db.ExecContext(ctx, "UPDATE tasks SET completed = CASE WHEN completed = 1 THEN 0 ELSE 1 END WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("failed to toggle completion: %w", err)
	}

	task, err := db.GetTaskByID(ctx, id)
	if err != nil {
		return false, err
	}
	return task.Completed, nil
}

func (db *DB) DeleteTask(ctx context.Context, id int64) error {
	res, err := db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
