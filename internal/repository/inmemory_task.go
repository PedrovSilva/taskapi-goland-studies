package repository

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pedrovsilva/taskapi/internal/domain"
)

// InMemoryTaskRepository is a thread-safe in-memory implementation of
// domain.TaskRepository. Used exclusively in tests — never in production.
type InMemoryTaskRepository struct {
	mu    sync.RWMutex
	tasks map[uuid.UUID]*domain.Task
}

// NewInMemoryTaskRepository creates an empty in-memory repository.
func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{
		tasks: make(map[uuid.UUID]*domain.Task),
	}
}

func (r *InMemoryTaskRepository) Create(task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *task
	r.tasks[task.ID] = &copy
	return nil
}

func (r *InMemoryTaskRepository) GetByID(id uuid.UUID) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task %s not found", id)
	}
	copy := *task
	return &copy, nil
}

func (r *InMemoryTaskRepository) List() ([]*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		copy := *t
		result = append(result, &copy)
	}
	return result, nil
}

func (r *InMemoryTaskRepository) Update(task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[task.ID]; !ok {
		return fmt.Errorf("task %s not found", task.ID)
	}
	copy := *task
	r.tasks[task.ID] = &copy
	return nil
}

func (r *InMemoryTaskRepository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[id]; !ok {
		return fmt.Errorf("task %s not found", id)
	}
	delete(r.tasks, id)
	return nil
}
