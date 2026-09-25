package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pedrovsilva/taskapi/internal/domain"
	"github.com/stretchr/testify/require"
)

func newPostgresTestRepository(t *testing.T) *PostgresTaskRepository {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	require.NoError(t, pool.Ping(ctx))

	_, err = pool.Exec(ctx, "TRUNCATE TABLE tasks")
	require.NoError(t, err)

	return NewPostgresTaskRepository(pool)
}

func newTestTask() *domain.Task {
	now := time.Now().UTC().Truncate(time.Microsecond)

	return &domain.Task{
		ID:          uuid.New(),
		Title:       "Integration task",
		Description: "PostgreSQL integration test",
		Status:      domain.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestPostgresTaskRepositoryCRUD(t *testing.T) {
	repo := newPostgresTestRepository(t)
	ctx := context.Background()
	task := newTestTask()

	// Create
	require.NoError(t, repo.Create(ctx, task))

	// GetByID
	got, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	require.Equal(t, task.ID, got.ID)
	require.Equal(t, task.Title, got.Title)
	require.Equal(t, task.Description, got.Description)
	require.Equal(t, task.Status, got.Status)

	require.WithinDuration(
		t,
		task.CreatedAt,
		got.CreatedAt,
		time.Millisecond,
	)

	// Update
	task.Title = "Updated task"
	task.Description = "Updated description"
	task.Status = domain.StatusInProgress
	task.UpdatedAt = time.Now().UTC()

	require.NoError(t, repo.Update(ctx, task))

	got, err = repo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	require.Equal(t, "Updated task", got.Title)
	require.Equal(t, "Updated description", got.Description)
	require.Equal(t, domain.StatusInProgress, got.Status)

	// List
	tasks, err := repo.List(ctx)
	require.NoError(t, err)

	require.Len(t, tasks, 1)
	require.Equal(t, task.ID, tasks[0].ID)

	// Delete
	require.NoError(t, repo.Delete(ctx, task.ID))

	_, err = repo.GetByID(ctx, task.ID)
	require.Error(t, err)
}

func TestPostgresTaskRepositoryPersistsDoneAt(t *testing.T) {
	repo := newPostgresTestRepository(t)
	ctx := context.Background()
	task := newTestTask()

	doneAt := time.Now().UTC()

	task.Status = domain.StatusDone
	task.DoneAt = &doneAt
	task.UpdatedAt = doneAt

	require.NoError(t, repo.Create(ctx, task))

	got, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	require.NotNil(t, got.DoneAt)

	require.WithinDuration(
		t,
		doneAt,
		*got.DoneAt,
		time.Millisecond,
	)
}

func TestPostgresTaskRepositoryNotFound(t *testing.T) {
	repo := newPostgresTestRepository(t)
	ctx := context.Background()
	id := uuid.New()

	// GetByID
	_, err := repo.GetByID(ctx, id)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")

	// Update
	task := newTestTask()
	task.ID = id

	require.Error(t, repo.Update(ctx, task))

	// Delete
	require.Error(t, repo.Delete(ctx, id))
}
