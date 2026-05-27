package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Task struct {
	ID          string
	Title       string
	Description string
	DueDate     string
	Done        bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(dbURL string) (*TaskRepository, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTable(db); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &TaskRepository{db: db}, nil
}

func createTable(db *sql.DB) error {
	query := `
        CREATE TABLE IF NOT EXISTS tasks (
            id VARCHAR(36) PRIMARY KEY,
            title VARCHAR(255) NOT NULL,
            description TEXT,
            due_date VARCHAR(50),
            done BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `
	_, err := db.Exec(query)
	return err
}

func (r *TaskRepository) Create(ctx context.Context, task *Task) error {
	query := `
        INSERT INTO tasks (id, title, description, due_date, done, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.DueDate,
		task.Done, task.CreatedAt, task.UpdatedAt,
	)
	return err
}

func (r *TaskRepository) GetAll(ctx context.Context) ([]*Task, error) {
	query := `
        SELECT id, title, description, due_date, done, created_at, updated_at
        FROM tasks ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.DueDate,
			&t.Done, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*Task, error) {
	query := `
        SELECT id, title, description, due_date, done, created_at, updated_at
        FROM tasks WHERE id = $1
    `
	var t Task
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.DueDate,
		&t.Done, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *Task) error {
	query := `
        UPDATE tasks
        SET title = $1, description = $2, due_date = $3, done = $4, updated_at = $5
        WHERE id = $6
    `
	_, err := r.db.ExecContext(ctx, query,
		task.Title, task.Description, task.DueDate, task.Done,
		task.UpdatedAt, task.ID,
	)
	return err
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func (r *TaskRepository) Close() error {
	return r.db.Close()
}
