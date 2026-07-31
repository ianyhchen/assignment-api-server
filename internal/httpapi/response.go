package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ianyhchen/assignment-api-server/internal/task"
)

type taskResponse struct {
	ID     uint64      `json:"id"`
	Name   string      `json:"name"`
	Status task.Status `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func newTaskResponse(model task.Task) taskResponse {
	return taskResponse{
		ID:     model.ID,
		Name:   model.Name,
		Status: model.Status,
	}
}

func newTaskResponses(models []task.Task) []taskResponse {
	responses := make([]taskResponse, len(models))

	for i, model := range models {
		responses[i] = newTaskResponse(model)
	}

	return responses
}

func writeJSON(w http.ResponseWriter, status int, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal JSON response: %w", err)
	}

	body = append(body, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("write JSON response: %w", err)
	}

	return nil
}

func writeError(w http.ResponseWriter, status int, message string) error {
	return writeJSON(w, status, errorResponse{
		Error: message,
	})
}
