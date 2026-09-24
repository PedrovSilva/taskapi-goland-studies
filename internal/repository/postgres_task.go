package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pedrovsilva/taskapi/internal/domain"
)

// PostgresTaskRepository implements domain.TaskRepository using PostgreSQL.
type PostgresTaskRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTaskRepository creates a new repository backed by the given connection pool.
func NewPostgresTaskRepository(pool *pgxpool.Pool) *PostgresTaskRepository {
	return &PostgresTaskRepository{pool: pool}
}

func (r *PostgresTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, title, description, status, created_at, updated_at, done_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		task.ID, task.Title, task.Description, task.Status,
		task.CreatedAt, task.UpdatedAt, task.DoneAt,
	)
	if err != nil {
		return fmt.Errorf("postgres: create task: %w", err)
	}
	return nil
}

func (r *PostgresTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	query := `
		SELECT id, title, description, status, created_at, updated_at, done_at
		FROM tasks WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	task := &domain.Task{}
	err := row.Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatedAt, &task.UpdatedAt, &task.DoneAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("task %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get task: %w", err)
	}
	return task, nil
}

func (r *PostgresTaskRepository) List(ctx context.Context) ([]*domain.Task, error) {
	query := `
		SELECT id, title, description, status, created_at, updated_at, done_at
		FROM tasks ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("postgres: list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		if err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status,
			&task.CreatedAt, &task.UpdatedAt, &task.DoneAt,
		); err != nil {
			return nil, fmt.Errorf("postgres: scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *PostgresTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, updated_at = $4, done_at = $5
		WHERE id = $6
	`
	result, err := r.pool.Exec(ctx, query,
		task.Title, task.Description, task.Status,
		task.UpdatedAt, task.DoneAt, task.ID,
	)
	if err != nil {
		return fmt.Errorf("postgres: update task: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("task %s not found", task.ID)
	}
	return nil
}

func (r *PostgresTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres: delete task: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("task %s not found", id)
	}
	return nil
}
