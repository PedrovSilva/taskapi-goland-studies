package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Status represents the current state of a task.
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// Task is the core domain entity.
type Task struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DoneAt      *time.Time `json:"done_at,omitempty"`
}

// Validate checks that the task is in a valid state.
func (t *Task) Validate() error {
	if t.Title == "" {
		return errors.New("title is required")
	}
	if len(t.Title) > 255 {
		return errors.New("title must be at most 255 characters")
	}
	return nil
}

// MarkDone transitions the task to done status.
func (t *Task) MarkDone() error {
	if t.Status == StatusDone {
		return errors.New("task is already done")
	}
	now := time.Now()
	t.Status = StatusDone
	t.DoneAt = &now
	t.UpdatedAt = now
	return nil
}

// TaskRepository defines the persistence contract for tasks.
// Any storage backend must implement this interface.
type TaskRepository interface {
	Create(task *Task) error
	GetByID(id uuid.UUID) (*Task, error)
	List() ([]*Task, error)
	Update(task *Task) error
	Delete(id uuid.UUID) error
}
