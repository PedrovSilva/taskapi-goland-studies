package service_test

import (
	"testing"

	"github.com/pedrovsilva/taskapi/internal/domain"
	"github.com/pedrovsilva/taskapi/internal/repository"
	"github.com/pedrovsilva/taskapi/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newService() *service.TaskService {
	return service.NewTaskService(repository.NewInMemoryTaskRepository())
}

func TestTaskService_Create(t *testing.T) {
	svc := newService()

	t.Run("creates task with valid input", func(t *testing.T) {
		task, err := svc.Create(service.CreateInput{
			Title:       "Write tests",
			Description: "Cover the happy path and error cases",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, task.ID)
		assert.Equal(t, "Write tests", task.Title)
		assert.Equal(t, domain.StatusPending, task.Status)
	})

	t.Run("returns error for empty title", func(t *testing.T) {
		_, err := svc.Create(service.CreateInput{Title: ""})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})
}

func TestTaskService_GetByID(t *testing.T) {
	svc := newService()

	task, err := svc.Create(service.CreateInput{Title: "Find me"})
	require.NoError(t, err)

	t.Run("returns task for existing id", func(t *testing.T) {
		found, err := svc.GetByID(task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.ID, found.ID)
	})

	t.Run("returns error for unknown id", func(t *testing.T) {
		_, err := svc.GetByID([16]byte{})
		require.Error(t, err)
	})
}

func TestTaskService_Update(t *testing.T) {
	svc := newService()
	task, _ := svc.Create(service.CreateInput{Title: "Original"})

	t.Run("updates title", func(t *testing.T) {
		newTitle := "Updated"
		updated, err := svc.Update(task.ID, service.UpdateInput{Title: &newTitle})
		require.NoError(t, err)
		assert.Equal(t, "Updated", updated.Title)
	})

	t.Run("transitions to done sets done_at", func(t *testing.T) {
		status := domain.StatusDone
		updated, err := svc.Update(task.ID, service.UpdateInput{Status: &status})
		require.NoError(t, err)
		assert.Equal(t, domain.StatusDone, updated.Status)
		assert.NotNil(t, updated.DoneAt)
	})

	t.Run("returns error for non-existent task", func(t *testing.T) {
		_, err := svc.Update([16]byte{}, service.UpdateInput{})
		require.Error(t, err)
	})
}

func TestTaskService_Delete(t *testing.T) {
	svc := newService()
	task, _ := svc.Create(service.CreateInput{Title: "Delete me"})

	t.Run("deletes existing task", func(t *testing.T) {
		err := svc.Delete(task.ID)
		require.NoError(t, err)

		_, err = svc.GetByID(task.ID)
		require.Error(t, err)
	})

	t.Run("returns error for non-existent task", func(t *testing.T) {
		err := svc.Delete([16]byte{})
		require.Error(t, err)
	})
}

func TestTaskService_List(t *testing.T) {
	svc := newService()

	t.Run("returns empty list when no tasks", func(t *testing.T) {
		tasks, err := svc.List()
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("returns all created tasks", func(t *testing.T) {
		svc.Create(service.CreateInput{Title: "Task A"})
		svc.Create(service.CreateInput{Title: "Task B"})

		tasks, err := svc.List()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(tasks), 2)
	})
}
