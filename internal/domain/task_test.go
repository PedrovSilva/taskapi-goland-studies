package domain_test

import (
	"testing"

	"github.com/pedrovsilva/taskapi/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask_Validate(t *testing.T) {
	t.Run("valid task passes validation", func(t *testing.T) {
		task := &domain.Task{Title: "Learn Go"}

		assert.NoError(t, task.Validate())
	})

	t.Run("empty title fails validation", func(t *testing.T) {
		task := &domain.Task{Title: ""}

		err := task.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("title with 255 characters is valid", func(t *testing.T) {
		task := &domain.Task{
			Title: string(make([]byte, 255)),
		}

		assert.NoError(t, task.Validate())
	})

	t.Run("title over 255 characters fails validation", func(t *testing.T) {
		task := &domain.Task{
			Title: string(make([]byte, 256)),
		}

		err := task.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "255 characters")
	})
}

func TestTask_MarkDone(t *testing.T) {
	t.Run("pending task can be marked done", func(t *testing.T) {
		task := &domain.Task{
			Title:  "Buy milk",
			Status: domain.StatusPending,
		}

		err := task.MarkDone()

		require.NoError(t, err)
		assert.Equal(t, domain.StatusDone, task.Status)
		assert.NotNil(t, task.DoneAt)
		assert.False(t, task.UpdatedAt.IsZero())
	})

	t.Run("in progress task can be marked done", func(t *testing.T) {
		task := &domain.Task{
			Title:  "Finish project",
			Status: domain.StatusInProgress,
		}

		err := task.MarkDone()

		require.NoError(t, err)
		assert.Equal(t, domain.StatusDone, task.Status)
		assert.NotNil(t, task.DoneAt)
	})

	t.Run("already done task returns error", func(t *testing.T) {
		task := &domain.Task{
			Title:  "Buy milk",
			Status: domain.StatusDone,
		}

		err := task.MarkDone()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already done")
	})
}

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.Status
		valid  bool
	}{
		{
			name:   "pending is valid",
			status: domain.StatusPending,
			valid:  true,
		},
		{
			name:   "in progress is valid",
			status: domain.StatusInProgress,
			valid:  true,
		},
		{
			name:   "done is valid",
			status: domain.StatusDone,
			valid:  true,
		},
		{
			name:   "unknown status is invalid",
			status: domain.Status("cancelled"),
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.status.IsValid())
		})
	}
}
