package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/pedrovsilva/taskapi/internal/domain"
	"github.com/pedrovsilva/taskapi/internal/handler"
	"github.com/pedrovsilva/taskapi/internal/repository"
	"github.com/pedrovsilva/taskapi/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRouter() http.Handler {
	repo := repository.NewInMemoryTaskRepository()
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		h.Routes(r)
	})
	return r
}

func TestHandler_CreateTask(t *testing.T) {
	router := newTestRouter()

	t.Run("POST /tasks returns 201 with valid body", func(t *testing.T) {
		body := `{"title":"Handler test task","description":"created via http"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var task domain.Task
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &task))
		assert.Equal(t, "Handler test task", task.Title)
		assert.Equal(t, domain.StatusPending, task.Status)
	})

	t.Run("POST /tasks returns 422 with empty title", func(t *testing.T) {
		body := `{"title":""}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestHandler_ListTasks(t *testing.T) {
	router := newTestRouter()

	// Seed a task
	body := `{"title":"Listed task"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	httptest.NewRecorder() // discard
	router.ServeHTTP(httptest.NewRecorder(), req)

	t.Run("GET /tasks returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandler_GetByID(t *testing.T) {
	router := newTestRouter()

	// Create a task first
	body := `{"title":"Get by ID test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var created domain.Task
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))

	t.Run("GET /tasks/:id returns 200 for existing task", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+created.ID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GET /tasks/:id returns 400 for invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/not-a-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_DeleteTask(t *testing.T) {
	router := newTestRouter()

	// Create
	body := `{"title":"Delete me"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var created domain.Task
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))

	t.Run("DELETE /tasks/:id returns 204", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+created.ID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestHandler_CreateTask_InvalidBody(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/tasks",
		bytes.NewBufferString(`{"title":`),
	)

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	router := newTestRouter()

	id := "550e8400-e29b-41d4-a716-446655440000"

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/tasks/"+id,
		nil,
	)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_UpdateTask(t *testing.T) {
	router := newTestRouter()

	// Create task.
	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/tasks",
		bytes.NewBufferString(`{"title":"Original"}`),
	)
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	require.Equal(t, http.StatusCreated, createW.Code)

	var created domain.Task
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &created))

	t.Run("PATCH updates task", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPatch,
			"/api/v1/tasks/"+created.ID.String(),
			bytes.NewBufferString(`{"title":"Updated"}`),
		)

		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updated domain.Task
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))

		assert.Equal(t, "Updated", updated.Title)
	})

	t.Run("PATCH rejects invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPatch,
			"/api/v1/tasks/"+created.ID.String(),
			bytes.NewBufferString(`{"title":`),
		)

		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PATCH rejects invalid status", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPatch,
			"/api/v1/tasks/"+created.ID.String(),
			bytes.NewBufferString(`{"status":"cancelled"}`),
		)

		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestHandler_DeleteTask_NotFound(t *testing.T) {
	router := newTestRouter()

	id := "550e8400-e29b-41d4-a716-446655440000"

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/tasks/"+id,
		nil,
	)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
