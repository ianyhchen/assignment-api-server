package task

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type fakeStore struct {
	listFunc   func(context.Context) ([]Task, error)
	createFunc func(context.Context, Task) (Task, error)
	updateFunc func(context.Context, Task) (Task, error)
	deleteFunc func(context.Context, uint64) error
}

func (store *fakeStore) Create(ctx context.Context, newTask Task) (Task, error) {
	if store.createFunc == nil {
		panic("unexpected call to Store.Create")
	}

	return store.createFunc(ctx, newTask)
}

func (store *fakeStore) List(ctx context.Context) ([]Task, error) {
	if store.listFunc == nil {
		panic("unexpected call to Store.List")
	}

	return store.listFunc(ctx)
}

func (store *fakeStore) Update(ctx context.Context, updatedTask Task) (Task, error) {
	if store.updateFunc == nil {
		panic("unexpected call to Store.Update")
	}
	return store.updateFunc(ctx, updatedTask)
}

func (store *fakeStore) Delete(ctx context.Context, id uint64) error {
	if store.deleteFunc == nil {
		panic("unexpected call to Store.Delete")
	}
	return store.deleteFunc(ctx, id)
}

func TestServiceCreate(t *testing.T) {
	var receivedTask Task

	store := &fakeStore{
		createFunc: func(_ context.Context, newTask Task) (Task, error) {
			// validate the data from service to store
			receivedTask = newTask
			newTask.ID = 1
			return newTask, nil
		},
	}

	service := NewService(store)

	created, err := service.Create(context.Background(), CreateInput{
		Name:   " Prepare test ",
		Status: StatusIncomplete,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	wantStoredTask := Task{
		Name:   "Prepare test",
		Status: StatusIncomplete,
	}
	if receivedTask != wantStoredTask {
		t.Errorf("Store.Create() task = %#v, want %#v", receivedTask, wantStoredTask)
	}

	wantCreatedTask := Task{
		ID:     1,
		Name:   "Prepare test",
		Status: StatusIncomplete,
	}
	if created != wantCreatedTask {
		t.Errorf("Create() task = %#v, want %#v", created, wantCreatedTask)
	}
}

func TestServiceCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateInput
		wantErr error
	}{
		{
			name: "empty name",
			input: CreateInput{
				Name:   "",
				Status: StatusIncomplete,
			},
			wantErr: ErrInvalidName,
		},
		{
			name: "whitespace-only name",
			input: CreateInput{
				Name:   "   ",
				Status: StatusIncomplete,
			},
			wantErr: ErrInvalidName,
		},
		{
			name: "negative status",
			input: CreateInput{
				Name:   "Task",
				Status: Status(-1),
			},
			wantErr: ErrInvalidStatus,
		},
		{
			name: "status above valid range",
			input: CreateInput{
				Name:   "Task",
				Status: Status(2),
			},
			wantErr: ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeCalled := false

			store := &fakeStore{
				createFunc: func(_ context.Context, newTask Task) (Task, error) {
					storeCalled = true
					return Task{}, nil
				},
			}

			service := NewService(store)

			created, err := service.Create(
				context.Background(),
				tt.input,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf(
					"Create() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}

			if created != (Task{}) {
				t.Errorf(
					"Create() task = %#v, want zero-value task",
					created,
				)
			}

			if storeCalled {
				t.Error("Store.Create() was called for invalid input")
			}
		})
	}
}

func TestServiceCreateWithCanceledContext(t *testing.T) {
	storeCalled := false

	store := &fakeStore{
		createFunc: func(_ context.Context, newTask Task) (Task, error) {
			storeCalled = true
			return Task{}, nil
		},
	}

	service := NewService(store)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.Create(ctx, CreateInput{
		Name:   "Task",
		Status: StatusIncomplete,
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Create() error = %v, want context.Canceled", err)
	}

	if storeCalled {
		t.Error("Store.Create() was called with canceled context")
	}
}

func TestServiceCreatePropagatesStoreError(t *testing.T) {
	storeErr := errors.New("storage unavailable")

	store := &fakeStore{
		createFunc: func(_ context.Context, newTask Task) (Task, error) {
			return Task{}, storeErr
		},
	}

	service := NewService(store)

	created, err := service.Create(context.Background(), CreateInput{
		Name:   "Task",
		Status: StatusIncomplete,
	})

	if !errors.Is(err, storeErr) {
		t.Errorf("Create() error = %v, want wrapped store error", err)
	}

	if created != (Task{}) {
		t.Errorf(
			"Create() task = %#v, want zero-value task",
			created,
		)
	}
}

func TestServiceList(t *testing.T) {
	wantTasks := []Task{
		{
			ID:     1,
			Name:   "First task",
			Status: StatusIncomplete,
		},
		{
			ID:     2,
			Name:   "Second task",
			Status: StatusCompleted,
		},
	}

	storeCalled := false

	store := &fakeStore{
		listFunc: func(_ context.Context) ([]Task, error) {
			storeCalled = true
			return wantTasks, nil
		},
	}

	service := NewService(store)

	gotTasks, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if !storeCalled {
		t.Error("Store.List() was not called")
	}

	if !slices.Equal(gotTasks, wantTasks) {
		t.Errorf(
			"List() tasks = %#v, want %#v",
			gotTasks,
			wantTasks,
		)
	}
}

func TestServiceListReturnsEmptySlice(t *testing.T) {
	store := &fakeStore{
		listFunc: func(_ context.Context) ([]Task, error) {
			return []Task{}, nil
		},
	}

	service := NewService(store)

	tasks, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if tasks == nil {
		t.Fatal("List() returned nil, want non-nil empty slice")
	}

	if len(tasks) != 0 {
		t.Errorf("List() returned %d tasks, want 0", len(tasks))
	}
}

func TestServiceListPropagatesStoreError(t *testing.T) {
	storeErr := errors.New("storage unavailable")

	store := &fakeStore{
		listFunc: func(_ context.Context) ([]Task, error) {
			return nil, storeErr
		},
	}

	service := NewService(store)

	tasks, err := service.List(context.Background())

	if !errors.Is(err, storeErr) {
		t.Errorf("List() error = %v, want wrapped store error", err)
	}

	if tasks != nil {
		t.Errorf(
			"List() tasks = %#v, want nil on error",
			tasks,
		)
	}
}

func TestServiceListWithCanceledContext(t *testing.T) {
	storeCalled := false

	store := &fakeStore{
		listFunc: func(_ context.Context) ([]Task, error) {
			storeCalled = true
			return nil, nil
		},
	}

	service := NewService(store)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tasks, err := service.List(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("List() error = %v, want context.Canceled", err)
	}

	if tasks != nil {
		t.Errorf(
			"List() tasks = %#v, want nil",
			tasks,
		)
	}

	if storeCalled {
		t.Error("Store.List() was called with canceled context")
	}
}

func TestServiceUpdate(t *testing.T) {
	var receivedTask Task

	store := &fakeStore{
		updateFunc: func(_ context.Context, updatedTask Task) (Task, error) {
			receivedTask = updatedTask
			return updatedTask, nil
		},
	}

	service := NewService(store)

	updated, err := service.Update(context.Background(),
		UpdateInput{
			ID:     42,
			Name:   "  Update task  ",
			Status: StatusCompleted,
		},
	)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	wantTask := Task{
		ID:     42,
		Name:   "Update task",
		Status: StatusCompleted,
	}

	if receivedTask != wantTask {
		t.Errorf("Store.Update() task = %#v, want %#v", receivedTask, wantTask)
	}

	if updated != wantTask {
		t.Errorf("Update() task = %#v, want %#v", updated, wantTask)
	}
}

func TestServiceUpdateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   UpdateInput
		wantErr error
	}{
		{
			name: "zero ID",
			input: UpdateInput{
				ID:     0,
				Name:   "Task",
				Status: StatusIncomplete,
			},
			wantErr: ErrInvalidTaskID,
		},
		{
			name: "empty name",
			input: UpdateInput{
				ID:     1,
				Name:   "",
				Status: StatusIncomplete,
			},
			wantErr: ErrInvalidName,
		},
		{
			name: "whitespace-only name",
			input: UpdateInput{
				ID:     1,
				Name:   "   ",
				Status: StatusIncomplete,
			},
			wantErr: ErrInvalidName,
		},
		{
			name: "negative status",
			input: UpdateInput{
				ID:     1,
				Name:   "Task",
				Status: Status(-1),
			},
			wantErr: ErrInvalidStatus,
		},
		{
			name: "status above valid range",
			input: UpdateInput{
				ID:     1,
				Name:   "Task",
				Status: Status(2),
			},
			wantErr: ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeCalled := false

			store := &fakeStore{
				updateFunc: func(context.Context, Task) (Task, error) {
					storeCalled = true
					return Task{}, nil
				},
			}

			service := NewService(store)

			updated, err := service.Update(
				context.Background(),
				tt.input,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf(
					"Update() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}

			if updated != (Task{}) {
				t.Errorf(
					"Update() task = %#v, want zero-value task",
					updated,
				)
			}

			if storeCalled {
				t.Error("Store.Update() was called for invalid input")
			}
		})
	}
}

func TestServiceUpdatePropagatesNotFound(t *testing.T) {
	store := &fakeStore{
		updateFunc: func(context.Context, Task) (Task, error) {
			return Task{}, ErrTaskNotFound
		},
	}

	service := NewService(store)

	updated, err := service.Update(
		context.Background(),
		UpdateInput{
			ID:     999,
			Name:   "Missing task",
			Status: StatusCompleted,
		},
	)

	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf(
			"Update() error = %v, want ErrTaskNotFound",
			err,
		)
	}

	if updated != (Task{}) {
		t.Errorf(
			"Update() task = %#v, want zero-value task",
			updated,
		)
	}
}

func TestServiceUpdateWithCanceledContext(t *testing.T) {
	storeCalled := false

	store := &fakeStore{
		updateFunc: func(context.Context, Task) (Task, error) {
			storeCalled = true
			return Task{}, nil
		},
	}

	service := NewService(store)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.Update(ctx, UpdateInput{
		ID:     1,
		Name:   "Task",
		Status: StatusCompleted,
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Update() error = %v, want context.Canceled", err)
	}

	if storeCalled {
		t.Error("Store.Update() was called with canceled context")
	}
}

func TestServiceDelete(t *testing.T) {
	var receivedID uint64
	storeCalled := false

	store := &fakeStore{
		deleteFunc: func(_ context.Context, id uint64) error {
			storeCalled = true
			receivedID = id
			return nil
		},
	}

	service := NewService(store)

	err := service.Delete(context.Background(), 42)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if !storeCalled {
		t.Fatal("Store.Delete() was not called")
	}

	if receivedID != 42 {
		t.Errorf(
			"Store.Delete() ID = %d, want 42",
			receivedID,
		)
	}
}

func TestServiceDeleteRejectsInvalidID(t *testing.T) {
	storeCalled := false

	store := &fakeStore{
		deleteFunc: func(context.Context, uint64) error {
			storeCalled = true
			return nil
		},
	}

	service := NewService(store)

	err := service.Delete(context.Background(), 0)

	if !errors.Is(err, ErrInvalidTaskID) {
		t.Errorf(
			"Delete() error = %v, want ErrInvalidTaskID",
			err,
		)
	}

	if storeCalled {
		t.Error("Store.Delete() was called with invalid ID")
	}
}

func TestServiceDeletePropagatesNotFound(t *testing.T) {
	store := &fakeStore{
		deleteFunc: func(context.Context, uint64) error {
			return ErrTaskNotFound
		},
	}

	service := NewService(store)

	err := service.Delete(context.Background(), 999)

	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf(
			"Delete() error = %v, want ErrTaskNotFound",
			err,
		)
	}
}

func TestServiceDeleteWithCanceledContext(t *testing.T) {
	storeCalled := false

	store := &fakeStore{
		deleteFunc: func(context.Context, uint64) error {
			storeCalled = true
			return nil
		},
	}

	service := NewService(store)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := service.Delete(ctx, 1)

	if !errors.Is(err, context.Canceled) {
		t.Errorf(
			"Delete() error = %v, want context.Canceled",
			err,
		)
	}

	if storeCalled {
		t.Error("Store.Delete() was called with canceled context")
	}
}
