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

	t.Run("title over 255 chars fails validation", func(t *testing.T) {
		task := &domain.Task{Title: string(make([]byte, 256))}
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
