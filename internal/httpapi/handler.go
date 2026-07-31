package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/ianyhchen/assignment-api-server/internal/task"
)

// TaskService defines the task operations required by the HTTP handlers.
type TaskService interface {
	List(ctx context.Context) ([]task.Task, error)
	Create(ctx context.Context, input task.CreateInput) (task.Task, error)
	Update(ctx context.Context, input task.UpdateInput) (task.Task, error)
	Delete(ctx context.Context, id uint64) error
}

// Handler handles HTTP requests for task operations.
type Handler struct {
	service TaskService
}

// NewHandler creates a Handler with the provided task service.
func NewHandler(service TaskService) *Handler {
	return &Handler{
		service: service,
	}
}

type createTaskRequest struct {
	Name   *string      `json:"name"`
	Status *task.Status `json:"status"`
}

type updateTaskRequest struct {
	Name   *string      `json:"name"`
	Status *task.Status `json:"status"`
}

func (handler *Handler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := handler.service.List(r.Context())
	if err != nil {
		if writeErr := writeTaskError(w, err); writeErr != nil {
			return
		}

		return
	}

	response := newTaskResponses(tasks)

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		return
	}
}

func (handler *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request createTaskRequest

	if err := decodeJSON(w, r, &request); err != nil {
		if writeErr := writeError(w, http.StatusBadRequest, err.Error()); writeErr != nil {
			return
		}

		return
	}

	if request.Name == nil {
		if err := writeError(w, http.StatusBadRequest, "name is required"); err != nil {
			return
		}

		return
	}

	if request.Status == nil {
		if err := writeError(w, http.StatusBadRequest, "status is required"); err != nil {
			return
		}

		return
	}

	created, err := handler.service.Create(
		r.Context(),
		task.CreateInput{
			Name:   *request.Name,
			Status: *request.Status,
		},
	)
	if err != nil {
		if writeErr := writeTaskError(w, err); writeErr != nil {
			return
		}

		return
	}

	w.Header().Set("Location", "/tasks/"+strconv.FormatUint(created.ID, 10))

	if err := writeJSON(w, http.StatusCreated, newTaskResponse(created)); err != nil {
		return
	}
}

func (handler *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r.PathValue("id"))
	if err != nil {
		if writeErr := writeTaskError(w, err); writeErr != nil {
			return
		}

		return
	}

	var request updateTaskRequest

	if err := decodeJSON(w, r, &request); err != nil {
		if writeErr := writeError(w, http.StatusBadRequest, err.Error()); writeErr != nil {
			return
		}

		return
	}

	if request.Name == nil {
		if err := writeError(w, http.StatusBadRequest, "name is required"); err != nil {
			return
		}

		return
	}

	if request.Status == nil {
		if err := writeError(w, http.StatusBadRequest, "status is required"); err != nil {
			return
		}

		return
	}

	updated, err := handler.service.Update(
		r.Context(),
		task.UpdateInput{
			ID:     id,
			Name:   *request.Name,
			Status: *request.Status,
		},
	)
	if err != nil {
		if writeErr := writeTaskError(w, err); writeErr != nil {
			return
		}

		return
	}

	if err := writeJSON(w, http.StatusOK, newTaskResponse(updated)); err != nil {
		return
	}
}

func (handler *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r.PathValue("id"))
	if err != nil {
		if writeErr := writeTaskError(w, err); writeErr != nil {
			return
		}

		return
	}

	if err := handler.service.Delete(r.Context(), id); err != nil {
		if writeErr := writeTaskError(w, err); writeErr != nil {
			return
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeTaskError(w http.ResponseWriter, err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		// The client canceled the request, so no response is written.
		return nil

	case errors.Is(err, context.DeadlineExceeded):
		return writeError(w, http.StatusGatewayTimeout, "request timed out")

	case errors.Is(err, task.ErrInvalidTaskID):
		return writeError(w, http.StatusBadRequest, task.ErrInvalidTaskID.Error())

	case errors.Is(err, task.ErrInvalidName):
		return writeError(w, http.StatusBadRequest, task.ErrInvalidName.Error())

	case errors.Is(err, task.ErrInvalidStatus):
		return writeError(w, http.StatusBadRequest, task.ErrInvalidStatus.Error())

	case errors.Is(err, task.ErrTaskNotFound):
		return writeError(w, http.StatusNotFound, task.ErrTaskNotFound.Error())

	default:
		return writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
