package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ianyhchen/assignment-api-server/internal/memory"
	"github.com/ianyhchen/assignment-api-server/internal/task"
)

func TestRouterTaskCRUDFlow(t *testing.T) {
	store := memory.NewTaskStore()
	service := task.NewService(store)
	handler := NewHandler(service)
	router := NewRouter(handler)

	// Create.
	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/tasks",
		strings.NewReader(`{"name":"Task","status":0}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")

	createRecorder := httptest.NewRecorder()
	router.ServeHTTP(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf(
			"POST /tasks status = %d, want %d; body = %s",
			createRecorder.Code,
			http.StatusCreated,
			createRecorder.Body.String(),
		)
	}

	var created taskResponse

	if err := json.NewDecoder(createRecorder.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	if created.ID != 1 {
		t.Errorf("created task ID = %d, want 1", created.ID)
	}

	// List after create.
	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/tasks", nil))

	if listRecorder.Code != http.StatusOK {
		t.Fatalf("GET /tasks status = %d, want %d", listRecorder.Code, http.StatusOK)
	}

	var listed []taskResponse

	if err := json.NewDecoder(listRecorder.Body).Decode(&listed); err != nil {
		t.Fatalf("decode list response: %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("GET /tasks returned %d tasks, want 1", len(listed))
	}

	if listed[0] != created {
		t.Errorf("listed task = %#v, want %#v", listed[0], created)
	}

	// Update.
	updateRequest := httptest.NewRequest(
		http.MethodPut,
		"/tasks/1",
		strings.NewReader(`{"name":"Updated task","status":1}`),
	)
	updateRequest.Header.Set("Content-Type", "application/json")

	updateRecorder := httptest.NewRecorder()
	router.ServeHTTP(updateRecorder, updateRequest)

	if updateRecorder.Code != http.StatusOK {
		t.Fatalf(
			"PUT /tasks/1 status = %d, want %d; body = %s",
			updateRecorder.Code,
			http.StatusOK,
			updateRecorder.Body.String(),
		)
	}

	var updated taskResponse

	if err := json.NewDecoder(updateRecorder.Body).Decode(&updated); err != nil {
		t.Fatalf("decode update response: %v", err)
	}

	if updated.Name != "Updated task" {
		t.Errorf("updated task name = %q, want %q", updated.Name, "Updated task")
	}

	if updated.Status != task.StatusCompleted {
		t.Errorf("updated status = %d, want %d", updated.Status, task.StatusCompleted)
	}

	// Delete.
	deleteRecorder := httptest.NewRecorder()
	router.ServeHTTP(deleteRecorder, httptest.NewRequest(http.MethodDelete, "/tasks/1", nil))

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("DELETE /tasks/1 status = %d, want %d", deleteRecorder.Code, http.StatusNoContent)
	}

	if deleteRecorder.Body.Len() != 0 {
		t.Errorf("DELETE response body = %q, want empty", deleteRecorder.Body.String())
	}

	// List after delete.
	finalListRecorder := httptest.NewRecorder()
	router.ServeHTTP(finalListRecorder, httptest.NewRequest(http.MethodGet, "/tasks", nil))

	if got := strings.TrimSpace(finalListRecorder.Body.String()); got != "[]" {
		t.Errorf("final GET /tasks body = %q, want %q", got, "[]")
	}
}
