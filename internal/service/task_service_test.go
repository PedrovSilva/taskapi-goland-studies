package service_test

import (
	"context"
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
		task, err := svc.Create(context.Background(), service.CreateInput{
			Title:       "Write tests",
			Description: "Cover the happy path and error cases",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, task.ID)
		assert.Equal(t, "Write tests", task.Title)
		assert.Equal(t, domain.StatusPending, task.Status)
	})

	t.Run("returns error for empty title", func(t *testing.T) {
		_, err := svc.Create(context.Background(), service.CreateInput{Title: ""})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})
}

func TestTaskService_GetByID(t *testing.T) {
	svc := newService()

	task, err := svc.Create(context.Background(), service.CreateInput{Title: "Find me"})
	require.NoError(t, err)

	t.Run("returns task for existing id", func(t *testing.T) {
		found, err := svc.GetByID(context.Background(), task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.ID, found.ID)
	})

	t.Run("returns error for unknown id", func(t *testing.T) {
		_, err := svc.GetByID(context.Background(), [16]byte{})
		require.Error(t, err)
	})
}

func TestTaskService_Update(t *testing.T) {
	t.Run("updates title without status", func(t *testing.T) {
		svc := newService()

		task, err := svc.Create(
			context.Background(),
			service.CreateInput{Title: "Original"},
		)
		require.NoError(t, err)

		newTitle := "Updated"

		updated, err := svc.Update(
			context.Background(),
			task.ID,
			service.UpdateInput{
				Title: &newTitle,
			},
		)

		require.NoError(t, err)
		assert.Equal(t, "Updated", updated.Title)
		assert.Equal(t, domain.StatusPending, updated.Status)
	})

	t.Run("updates description", func(t *testing.T) {
		svc := newService()

		task, err := svc.Create(
			context.Background(),
			service.CreateInput{Title: "Task"},
		)
		require.NoError(t, err)

		description := "Updated description"

		updated, err := svc.Update(
			context.Background(),
			task.ID,
			service.UpdateInput{
				Description: &description,
			},
		)

		require.NoError(t, err)
		assert.Equal(t, description, updated.Description)
	})

	t.Run("transitions pending to in progress", func(t *testing.T) {
		svc := newService()

		task, err := svc.Create(
			context.Background(),
			service.CreateInput{Title: "Task"},
		)
		require.NoError(t, err)

		status := domain.StatusInProgress

		updated, err := svc.Update(
			context.Background(),
			task.ID,
			service.UpdateInput{
				Status: &status,
			},
		)

		require.NoError(t, err)
		assert.Equal(t, domain.StatusInProgress, updated.Status)
	})

	t.Run("transitions to done and sets done_at", func(t *testing.T) {
		svc := newService()

		task, err := svc.Create(
			context.Background(),
			service.CreateInput{Title: "Task"},
		)
		require.NoError(t, err)

		status := domain.StatusDone

		updated, err := svc.Update(
			context.Background(),
			task.ID,
			service.UpdateInput{
				Status: &status,
			},
		)

		require.NoError(t, err)
		assert.Equal(t, domain.StatusDone, updated.Status)
		assert.NotNil(t, updated.DoneAt)
	})

	t.Run("rejects invalid status", func(t *testing.T) {
		svc := newService()

		task, err := svc.Create(
			context.Background(),
			service.CreateInput{Title: "Task"},
		)
		require.NoError(t, err)

		status := domain.Status("cancelled")

		_, err = svc.Update(
			context.Background(),
			task.ID,
			service.UpdateInput{
				Status: &status,
			},
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task status")
	})

	t.Run("cannot move done task back to pending", func(t *testing.T) {
		svc := newService()

		task, err := svc.Create(
			context.Background(),
			service.CreateInput{Title: "Task"},
		)
		require.NoError(t, err)

		done := domain.StatusDone

		_, err = svc.Update(
			context.Background(),
			task.ID,
			service.UpdateInput{
				Status: &done,
			},
		)
		require.NoError(t, err)

		pending := domain.StatusPending

		_, err = svc.Update(
			context.Background(),
			task.ID,
			service.UpdateInput{
				Status: &pending,
			},
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already marked done")
	})

	t.Run("cannot mark already done task as done", func(t *testing.T) {
		svc := newService()

		task, err := svc.Create(
			context.Background(),
			service.CreateInput{Title: "Task"},
		)
		require.NoError(t, err)

		done := domain.StatusDone

		_, err = svc.Update(
			context.Background(),
			task.ID,
			service.UpdateInput{
				Status: &done,
			},
		)
		require.NoError(t, err)

		_, err = svc.Update(
			context.Background(),
			task.ID,
			service.UpdateInput{
				Status: &done,
			},
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already done")
	})
}

func TestTaskService_Delete(t *testing.T) {
	svc := newService()
	task, _ := svc.Create(context.Background(), service.CreateInput{Title: "Delete me"})

	t.Run("deletes existing task", func(t *testing.T) {
		err := svc.Delete(context.Background(), task.ID)
		require.NoError(t, err)

		_, err = svc.GetByID(context.Background(), task.ID)
		require.Error(t, err)
	})

	t.Run("returns error for non-existent task", func(t *testing.T) {
		err := svc.Delete(context.Background(), [16]byte{})
		require.Error(t, err)
	})
}

func TestTaskService_List(t *testing.T) {
	svc := newService()

	t.Run("returns empty list when no tasks", func(t *testing.T) {
		tasks, err := svc.List(context.Background())
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("returns all created tasks", func(t *testing.T) {
		svc.Create(context.Background(), service.CreateInput{Title: "Task A"})
		svc.Create(context.Background(), service.CreateInput{Title: "Task B"})

		tasks, err := svc.List(context.Background())
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(tasks), 2)
	})
}
