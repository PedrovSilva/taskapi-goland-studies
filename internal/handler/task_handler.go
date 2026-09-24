package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pedrovsilva/taskapi/internal/service"
)

// TaskHandler exposes HTTP endpoints for the task resource.
type TaskHandler struct {
	svc *service.TaskService
}

// NewTaskHandler constructs a TaskHandler with the given service.
func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// Routes registers all task routes on the given router.
func (h *TaskHandler) Routes(r chi.Router) {
	r.Post("/tasks", h.Create)
	r.Get("/tasks", h.List)
	r.Get("/tasks/{id}", h.GetByID)
	r.Patch("/tasks/{id}", h.Update)
	r.Delete("/tasks/{id}", h.Delete)
}

// Create godoc
// @Summary      Create a task
// @Description  Creates a new task with pending status
// @Tags         tasks
// @Accept       JSON
// @Produce      JSON
// @Param        task  body      service.CreateInput  true  "Task input"
// @Success      201   {object}  domain.Task
// @Failure      400   {object}  errorResponse
// @Failure      422   {object}  errorResponse
// @Router       /tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.svc.Create(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, task)
}

// List godoc
// @Summary      List all tasks
// @Description  Returns all tasks ordered by creation date
// @Tags         tasks
// @Produce      JSON
// @Success      200  {array}   domain.Task
// @Failure      500  {object}  errorResponse
// @Router       /tasks [get]
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.svc.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tasks)
}

// GetByID godoc
// @Summary      Get task by ID
// @Description  Returns a single task by its UUID
// @Tags         tasks
// @Produce      JSON
// @Param        id   path      string  true  "Task UUID"
// @Success      200  {object}  domain.Task
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, task)
}

// Update godoc
// @Summary      Update a task
// @Description  Partially updates a task's title, description or status
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id    path      string               true  "Task UUID"
// @Param        task  body      service.UpdateInput  true  "Fields to update"
// @Success      200   {object}  domain.Task
// @Failure      400   {object}  errorResponse
// @Failure      422   {object}  errorResponse
// @Router       /tasks/{id} [patch]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var input service.UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, task)
}

// Delete godoc
// @Summary      Delete a task
// @Description  Permanently removes a task by its UUID
// @Tags         tasks
// @Param        id  path  string  true  "Task UUID"
// @Success      204
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// helpers

type errorResponse struct {
	Error string `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, errorResponse{Error: msg})
}
