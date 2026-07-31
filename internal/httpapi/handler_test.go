package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ianyhchen/assignment-api-server/internal/task"
)

type fakeTaskService struct {
	listFunc   func(context.Context) ([]task.Task, error)
	createFunc func(context.Context, task.CreateInput) (task.Task, error)
	updateFunc func(context.Context, task.UpdateInput) (task.Task, error)
	deleteFunc func(context.Context, uint64) error
}

func (service *fakeTaskService) List(ctx context.Context) ([]task.Task, error) {
	if service.listFunc == nil {
		panic("unexpected call to TaskService.List")
	}
	return service.listFunc(ctx)
}

func (service *fakeTaskService) Create(
	ctx context.Context,
	input task.CreateInput,
) (task.Task, error) {
	if service.createFunc == nil {
		panic("unexpected call to TaskService.Create")
	}

	return service.createFunc(ctx, input)
}

func (service *fakeTaskService) Update(
	ctx context.Context,
	input task.UpdateInput,
) (task.Task, error) {
	if service.updateFunc == nil {
		panic("unexpected call to TaskService.Update")
	}

	return service.updateFunc(ctx, input)
}

func (service *fakeTaskService) Delete(ctx context.Context, id uint64) error {
	if service.deleteFunc == nil {
		panic("unexpected call to TaskService.Delete")
	}

	return service.deleteFunc(ctx, id)
}

func TestHandlerList(t *testing.T) {
	service := &fakeTaskService{
		listFunc: func(context.Context) ([]task.Task, error) {
			return []task.Task{
				{
					ID:     1,
					Name:   "First task",
					Status: task.StatusIncomplete,
				},
				{
					ID:     2,
					Name:   "Second task",
					Status: task.StatusCompleted,
				},
			}, nil
		},
	}

	handler := NewHandler(service)

	request := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	if got := response.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	var got []taskResponse

	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	want := []taskResponse{
		{
			ID:     1,
			Name:   "First task",
			Status: task.StatusIncomplete,
		},
		{
			ID:     2,
			Name:   "Second task",
			Status: task.StatusCompleted,
		},
	}

	if len(got) != len(want) {
		t.Fatalf("response contains %d tasks, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf(
				"response[%d] = %#v, want %#v",
				i,
				got[i],
				want[i],
			)
		}
	}
}

func TestHandlerListReturnsEmptyArray(t *testing.T) {
	service := &fakeTaskService{
		listFunc: func(context.Context) ([]task.Task, error) {
			return nil, nil
		},
	}

	handler := NewHandler(service)

	request := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if got := strings.TrimSpace(recorder.Body.String()); got != "[]" {
		t.Errorf("response body = %q, want %q", got, "[]")
	}
}

func TestHandlerListDoesNotExposeInternalError(t *testing.T) {
	internalErr := errors.New("database password was rejected")

	service := &fakeTaskService{
		listFunc: func(context.Context) ([]task.Task, error) {
			return nil, internalErr
		},
	}

	handler := NewHandler(service)

	request := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	body := recorder.Body.Bytes()
	var got errorResponse

	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Error != "internal server error" {
		t.Errorf("error = %q, want %q", got.Error, "internal server error")
	}

	if strings.Contains(string(body), internalErr.Error()) {
		t.Error("response exposed the internal service error")
	}
}

func TestHandlerCreate(t *testing.T) {
	var receivedInput task.CreateInput

	service := &fakeTaskService{
		createFunc: func(_ context.Context, input task.CreateInput) (task.Task, error) {
			receivedInput = input

			return task.Task{
				ID:     1,
				Name:   input.Name,
				Status: input.Status,
			}, nil
		},
	}

	handler := NewHandler(service)

	request := httptest.NewRequest(
		http.MethodPost,
		"/tasks",
		strings.NewReader(`{"name":"Prepare assignment","status":0}`),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}

	if got := recorder.Header().Get("Location"); got != "/tasks/1" {
		t.Errorf("Location = %q, want %q", got, "/tasks/1")
	}

	wantInput := task.CreateInput{
		Name:   "Prepare assignment",
		Status: task.StatusIncomplete,
	}

	if receivedInput != wantInput {
		t.Errorf("TaskService.Create() input = %#v, want %#v", receivedInput, wantInput)
	}

	var got taskResponse

	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	wantResponse := taskResponse{
		ID:     1,
		Name:   "Prepare assignment",
		Status: task.StatusIncomplete,
	}

	if got != wantResponse {
		t.Errorf("response = %#v, want %#v", got, wantResponse)
	}
}

func TestHandlerCreateRequiresFields(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{
			name:      "missing name",
			body:      `{"status":0}`,
			wantError: "name is required",
		},
		{
			name:      "missing status",
			body:      `{"name":"Task"}`,
			wantError: "status is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false

			service := &fakeTaskService{
				createFunc: func(context.Context, task.CreateInput) (task.Task, error) {
					serviceCalled = true
					return task.Task{}, nil
				},
			}

			handler := NewHandler(service)

			request := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()

			handler.Create(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}

			var got errorResponse

			if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if got.Error != tt.wantError {
				t.Errorf("error = %q, want %q", got.Error, tt.wantError)
			}

			if serviceCalled {
				t.Error("TaskService.Create() was called for invalid request")
			}
		})
	}
}

func TestHandlerCreateMapsValidationError(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
	}{
		{
			name:       "invalid name",
			serviceErr: task.ErrInvalidName,
		},
		{
			name:       "invalid status",
			serviceErr: task.ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeTaskService{
				createFunc: func(context.Context, task.CreateInput) (task.Task, error) {
					return task.Task{}, tt.serviceErr
				},
			}

			handler := NewHandler(service)
			request := httptest.NewRequest(
				http.MethodPost,
				"/tasks",
				strings.NewReader(`{"name":"Task","status":0}`),
			)
			recorder := httptest.NewRecorder()

			handler.Create(recorder, request)

			assertErrorResponse(t, recorder, http.StatusBadRequest, tt.serviceErr.Error())
		})
	}
}

func TestHandlerCreateRejectsInvalidJSON(t *testing.T) {
	handler := NewHandler(&fakeTaskService{})
	request := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"name":`))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "request body contains invalid JSON")
}

func TestHandlerUpdate(t *testing.T) {
	var receivedInput task.UpdateInput

	service := &fakeTaskService{
		updateFunc: func(_ context.Context, input task.UpdateInput) (task.Task, error) {
			receivedInput = input

			return task.Task{
				ID:     input.ID,
				Name:   input.Name,
				Status: input.Status,
			}, nil
		},
	}

	handler := NewHandler(service)

	request := httptest.NewRequest(
		http.MethodPut,
		"/tasks/42",
		strings.NewReader(`{"name":"Updated task","status":1}`),
	)
	request.SetPathValue("id", "42")
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Update(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	wantInput := task.UpdateInput{
		ID:     42,
		Name:   "Updated task",
		Status: task.StatusCompleted,
	}

	if receivedInput != wantInput {
		t.Errorf("TaskService.Update() input = %#v, want %#v", receivedInput, wantInput)
	}

	var got taskResponse

	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	wantResponse := taskResponse{
		ID:     42,
		Name:   "Updated task",
		Status: task.StatusCompleted,
	}

	if got != wantResponse {
		t.Errorf("response = %#v, want %#v", got, wantResponse)
	}
}

func TestHandlerUpdateRequiresFields(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{
			name:      "missing name",
			body:      `{"status":0}`,
			wantError: "name is required",
		},
		{
			name:      "missing status",
			body:      `{"name":"Task"}`,
			wantError: "status is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(&fakeTaskService{})
			request := httptest.NewRequest(http.MethodPut, "/tasks/42", strings.NewReader(tt.body))
			request.SetPathValue("id", "42")
			recorder := httptest.NewRecorder()

			handler.Update(recorder, request)

			assertErrorResponse(t, recorder, http.StatusBadRequest, tt.wantError)
		})
	}
}

func TestHandlerUpdateRejectsInvalidJSON(t *testing.T) {
	handler := NewHandler(&fakeTaskService{})
	request := httptest.NewRequest(http.MethodPut, "/tasks/42", strings.NewReader(`{"name":`))
	request.SetPathValue("id", "42")
	recorder := httptest.NewRecorder()

	handler.Update(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "request body contains invalid JSON")
}

func TestHandlerUpdateRejectsInvalidID(t *testing.T) {
	tests := []string{
		"",
		"0",
		"-1",
		"abc",
	}

	for _, id := range tests {
		t.Run(id, func(t *testing.T) {
			serviceCalled := false

			service := &fakeTaskService{
				updateFunc: func(context.Context, task.UpdateInput) (task.Task, error) {
					serviceCalled = true
					return task.Task{}, nil
				},
			}

			handler := NewHandler(service)

			request := httptest.NewRequest(
				http.MethodPut,
				"/tasks/"+id,
				strings.NewReader(`{"name":"Task","status":1}`),
			)
			request.SetPathValue("id", id)

			recorder := httptest.NewRecorder()

			handler.Update(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}

			if serviceCalled {
				t.Error("TaskService.Update() was called with invalid ID")
			}
		})
	}
}

func TestHandlerUpdateNotFound(t *testing.T) {
	service := &fakeTaskService{
		updateFunc: func(context.Context, task.UpdateInput) (task.Task, error) {
			return task.Task{}, task.ErrTaskNotFound
		},
	}

	handler := NewHandler(service)

	request := httptest.NewRequest(
		http.MethodPut,
		"/tasks/999",
		strings.NewReader(`{"name":"Missing task","status":1}`),
	)
	request.SetPathValue("id", "999")

	recorder := httptest.NewRecorder()

	handler.Update(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}

	var got errorResponse

	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Error != task.ErrTaskNotFound.Error() {
		t.Errorf("error = %q, want %q", got.Error, task.ErrTaskNotFound.Error())
	}
}

func TestHandlerDelete(t *testing.T) {
	var receivedID uint64

	service := &fakeTaskService{
		deleteFunc: func(_ context.Context, id uint64) error {
			receivedID = id
			return nil
		},
	}

	handler := NewHandler(service)

	request := httptest.NewRequest(http.MethodDelete, "/tasks/42", nil)
	request.SetPathValue("id", "42")

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	if receivedID != 42 {
		t.Errorf("TaskService.Delete() ID = %d, want 42", receivedID)
	}

	if recorder.Body.Len() != 0 {
		t.Errorf("response body = %q, want empty body", recorder.Body.String())
	}
}

func TestHandlerDeleteNotFound(t *testing.T) {
	service := &fakeTaskService{
		deleteFunc: func(context.Context, uint64) error {
			return task.ErrTaskNotFound
		},
	}

	handler := NewHandler(service)

	request := httptest.NewRequest(http.MethodDelete, "/tasks/999", nil)
	request.SetPathValue("id", "999")

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestHandlerDeleteRejectsInvalidID(t *testing.T) {
	tests := []string{
		"",
		"0",
		"-1",
		"abc",
	}

	for _, id := range tests {
		t.Run(id, func(t *testing.T) {
			serviceCalled := false

			service := &fakeTaskService{
				deleteFunc: func(_ context.Context, id uint64) error {
					serviceCalled = true
					return nil
				},
			}

			handler := NewHandler(service)

			request := httptest.NewRequest(http.MethodDelete, "/tasks/"+id, nil)

			request.SetPathValue("id", id)

			recorder := httptest.NewRecorder()

			handler.Delete(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}

			if serviceCalled {
				t.Error("TaskService.Delete() was called with invalid ID")
			}
		})
	}
}

func assertErrorResponse(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	wantStatus int,
	wantMessage string,
) {
	t.Helper()

	if recorder.Code != wantStatus {
		t.Errorf("status = %d, want %d", recorder.Code, wantStatus)
	}

	var got errorResponse

	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Error != wantMessage {
		t.Errorf("error = %q, want %q", got.Error, wantMessage)
	}
}
