package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pedrovsilva/taskapi/internal/domain"
)

// TaskService holds the business logic for task operations.
// It depends on the TaskRepository interface, not on any concrete implementation.
type TaskService struct {
	repo domain.TaskRepository
}

// NewTaskService constructs a TaskService with the given repository.
func NewTaskService(repo domain.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

// CreateInput holds the data needed to create a new task.
type CreateInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateInput holds the fields that can be updated on a task.
type UpdateInput struct {
	Title       *string        `json:"title,omitempty"`
	Description *string        `json:"description,omitempty"`
	Status      *domain.Status `json:"status,omitempty"`
}

// Create validates and persists a new task.
func (s *TaskService) Create(ctx context.Context, input CreateInput) (*domain.Task, error) {
	now := time.Now().UTC()
	task := &domain.Task{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		Status:      domain.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("invalid task: %w", err)
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("could not create task: %w", err)
	}

	return task, nil
}

// GetByID retrieves a single task by its ID.
func (s *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not get task: %w", err)
	}
	return task, nil
}

// List retrieves all tasks.
func (s *TaskService) List(ctx context.Context) ([]*domain.Task, error) {
	tasks, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list tasks: %w", err)
	}
	return tasks, nil
}

// Update applies partial changes to an existing task.
func (s *TaskService) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	if input.Status != nil && !input.Status.IsValid() {
		return nil, fmt.Errorf("invalid task status: %s", *input.Status)
	}

	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Status != nil {
		if *input.Status == domain.StatusDone {
			if err := task.MarkDone(); err != nil {
				return nil, err
			}
		} else {
			if task.Status == domain.StatusDone {
				return nil, fmt.Errorf("task already marked done")
			}
			task.Status = *input.Status
		}
	}

	task.UpdatedAt = time.Now().UTC()

	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("invalid task: %w", err)
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("could not update task: %w", err)
	}

	return task, nil
}

// Delete removes a task by ID.
func (s *TaskService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("could not delete task: %w", err)
	}
	return nil
}
