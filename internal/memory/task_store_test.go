package memory

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/ianyhchen/assignment-api-server/internal/task"
)

func TestTaskStoreCreate(t *testing.T) {
	store := NewTaskStore()

	input := task.Task{
		Name:   "Test store create",
		Status: task.StatusIncomplete,
	}

	created, err := store.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID != 1 {
		t.Errorf("Create() ID = %d, want 1", created.ID)
	}

	if created.Name != input.Name {
		t.Errorf("Create() Name = %q, want %q", created.Name, input.Name)
	}

	if created.Status != input.Status {
		t.Errorf("Create() Status = %d, want %d", created.Status, input.Status)
	}
}

func TestTaskStoreCreateAssignsSequentialIDs(t *testing.T) {
	store := NewTaskStore()
	ctx := context.Background()

	first, err := store.Create(ctx, task.Task{
		Name:   "First task",
		Status: task.StatusIncomplete,
	})

	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	second, err := store.Create(ctx, task.Task{
		Name:   "Second task",
		Status: task.StatusCompleted,
	})

	if err != nil {
		t.Fatalf("second Create() error = %v", err)
	}

	if first.ID != 1 {
		t.Errorf("first task ID = %d, want 1", first.ID)
	}

	if second.ID != 2 {
		t.Errorf("second task ID = %d, want 2", second.ID)
	}
}

func TestTaskStoreCreateShouldIgnoresProvidedID(t *testing.T) {
	store := NewTaskStore()

	created, err := store.Create(context.Background(), task.Task{
		ID:     999,
		Name:   "Task",
		Status: task.StatusIncomplete,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID != 1 {
		t.Errorf("Create() ID = %d, want server-generated ID 1", created.ID)
	}
}

func TestTaskStoreCreateWithCanceledContext(t *testing.T) {
	store := NewTaskStore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.Create(ctx, task.Task{
		Name:   "Task",
		Status: task.StatusIncomplete,
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Create() error = %v, want context.Canceled", err)
	}
}

func TestTaskStoreListInitiallyEmpty(t *testing.T) {
	store := NewTaskStore()

	tasks, err := store.List(context.Background())
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

func TestTaskStoreListReturnsCreatedTasks(t *testing.T) {
	store := NewTaskStore()
	ctx := context.Background()

	created, err := store.Create(ctx, task.Task{
		Name:   "Prepare test",
		Status: task.StatusIncomplete,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	tasks, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("List() returned %d tasks, want 1", len(tasks))
	}

	if tasks[0] != created {
		t.Errorf("List()[0] = %#v, want %#v", tasks[0], created)
	}
}

func TestTaskStoreListSortsTasksByID(t *testing.T) {
	store := NewTaskStore()
	ctx := context.Background()

	for _, name := range []string{"First", "Second", "Third"} {
		_, err := store.Create(ctx, task.Task{
			Name:   name,
			Status: task.StatusIncomplete,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	tasks, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("List() returned %d tasks, want 3", len(tasks))
	}

	for i, storedTask := range tasks {
		wantID := uint64(i + 1)

		if storedTask.ID != wantID {
			t.Errorf(
				"tasks[%d].ID = %d, want %d",
				i,
				storedTask.ID,
				wantID,
			)
		}
	}
}

func TestTaskStoreListReturnsCopies(t *testing.T) {
	store := NewTaskStore()
	ctx := context.Background()

	_, err := store.Create(ctx, task.Task{
		Name:   "Original",
		Status: task.StatusIncomplete,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	firstResult, err := store.List(ctx)
	if err != nil {
		t.Fatalf("first List() error = %v", err)
	}

	firstResult[0].Name = "Changed outside store"

	secondResult, err := store.List(ctx)
	if err != nil {
		t.Fatalf("second List() error = %v", err)
	}

	if secondResult[0].Name != "Original" {
		t.Errorf(
			"stored task name = %q, want %q",
			secondResult[0].Name,
			"Original",
		)
	}
}

func TestTaskStoreListWithCanceledContext(t *testing.T) {
	store := NewTaskStore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tasks, err := store.List(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("List() error = %v, want context.Canceled", err)
	}

	if tasks != nil {
		t.Errorf("List() tasks = %#v, want nil on error", tasks)
	}
}

func TestTaskStoreUpdate(t *testing.T) {
	store := NewTaskStore()
	ctx := context.Background()

	created, err := store.Create(ctx, task.Task{
		Name:   "Original name",
		Status: task.StatusIncomplete,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := store.Update(ctx, task.Task{
		ID:     created.ID,
		Name:   "Updated name",
		Status: task.StatusCompleted,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.ID != created.ID {
		t.Errorf("Update() ID = %d, want %d", updated.ID, created.ID)
	}

	if updated.Name != "Updated name" {
		t.Errorf(
			"Update() Name = %q, want %q",
			updated.Name,
			"Updated name",
		)
	}

	if updated.Status != task.StatusCompleted {
		t.Errorf(
			"Update() Status = %d, want %d",
			updated.Status,
			task.StatusCompleted,
		)
	}

	tasks, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("List() returned %d tasks, want 1", len(tasks))
	}

	if tasks[0] != updated {
		t.Errorf("stored task = %#v, want %#v", tasks[0], updated)
	}
}

func TestTaskStoreUpdateNotFound(t *testing.T) {
	store := NewTaskStore()

	updated, err := store.Update(context.Background(), task.Task{
		ID:     999,
		Name:   "Missing task",
		Status: task.StatusCompleted,
	})

	if !errors.Is(err, task.ErrTaskNotFound) {
		t.Errorf(
			"Update() error = %v, want task.ErrTaskNotFound",
			err,
		)
	}

	if updated != (task.Task{}) {
		t.Errorf(
			"Update() task = %#v, want zero-value task",
			updated,
		)
	}

	tasks, listErr := store.List(context.Background())
	if listErr != nil {
		t.Fatalf("List() error = %v", listErr)
	}

	if len(tasks) != 0 {
		t.Errorf("List() returned %d tasks, want 0", len(tasks))
	}
}

func TestTaskStoreUpdateWithCanceledContext(t *testing.T) {
	store := NewTaskStore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.Update(ctx, task.Task{
		ID:     1,
		Name:   "Task",
		Status: task.StatusCompleted,
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Update() error = %v, want context.Canceled", err)
	}
}

func TestTaskStoreDelete(t *testing.T) {
	store := NewTaskStore()
	ctx := context.Background()

	created, err := store.Create(ctx, task.Task{
		Name:   "Task to delete",
		Status: task.StatusIncomplete,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	tasks, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("List() returned %d tasks, want 0", len(tasks))
	}
}

func TestTaskStoreDeleteOnlyRemovesRequestedTask(t *testing.T) {
	store := NewTaskStore()
	ctx := context.Background()

	first, err := store.Create(ctx, task.Task{
		Name:   "First task",
		Status: task.StatusIncomplete,
	})
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	second, err := store.Create(ctx, task.Task{
		Name:   "Second task",
		Status: task.StatusCompleted,
	})
	if err != nil {
		t.Fatalf("second Create() error = %v", err)
	}

	if err := store.Delete(ctx, first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	tasks, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("List() returned %d tasks, want 1", len(tasks))
	}

	if tasks[0] != second {
		t.Errorf("remaining task = %#v, want %#v", tasks[0], second)
	}
}

func TestTaskStoreDeleteNotFound(t *testing.T) {
	store := NewTaskStore()

	err := store.Delete(context.Background(), 999)

	if !errors.Is(err, task.ErrTaskNotFound) {
		t.Errorf(
			"Delete() error = %v, want task.ErrTaskNotFound",
			err,
		)
	}
}

func TestTaskStoreDeleteTwiceReturnsNotFound(t *testing.T) {
	store := NewTaskStore()
	ctx := context.Background()

	created, err := store.Create(ctx, task.Task{
		Name:   "Task",
		Status: task.StatusIncomplete,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("first Delete() error = %v", err)
	}

	err = store.Delete(ctx, created.ID)
	if !errors.Is(err, task.ErrTaskNotFound) {
		t.Errorf(
			"second Delete() error = %v, want task.ErrTaskNotFound",
			err,
		)
	}
}

func TestTaskStoreDeleteWithCanceledContext(t *testing.T) {
	store := NewTaskStore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := store.Delete(ctx, 1)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Delete() error = %v, want context.Canceled", err)
	}
}

func TestTaskStoreConcurrentAccess(t *testing.T) {
	const (
		writerCount = 100
		readerCount = 10
		readCount   = 100
	)

	store := NewTaskStore()
	ctx := context.Background()

	start := make(chan struct{})
	createdTasks := make(chan task.Task, writerCount)
	errs := make(chan error, writerCount+readerCount)

	var wg sync.WaitGroup

	// Concurrent writers exercise both nextID and the task map.
	for range writerCount {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// Wait until every goroutine has been created.
			<-start

			created, err := store.Create(ctx, task.Task{
				Name:   "Concurrent task",
				Status: task.StatusIncomplete,
			})
			if err != nil {
				errs <- err
				return
			}

			createdTasks <- created
		}()
	}

	// Concurrent readers exercise map iteration while writes are happening.
	for range readerCount {
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-start

			for range readCount {
				if _, err := store.List(ctx); err != nil {
					errs <- err
					return
				}
			}
		}()
	}

	// Release all goroutines at approximately the same time.
	close(start)

	wg.Wait()
	close(createdTasks)
	close(errs)

	for err := range errs {
		t.Errorf("concurrent store operation returned error: %v", err)
	}

	seenIDs := make(map[uint64]struct{}, writerCount)

	for created := range createdTasks {
		if created.ID == 0 {
			t.Error("Create() returned ID 0")
			continue
		}

		if _, exists := seenIDs[created.ID]; exists {
			t.Errorf("Create() returned duplicate ID %d", created.ID)
			continue
		}

		seenIDs[created.ID] = struct{}{}
	}

	if len(seenIDs) != writerCount {
		t.Errorf(
			"Create() produced %d unique IDs, want %d",
			len(seenIDs),
			writerCount,
		)
	}

	storedTasks, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(storedTasks) != writerCount {
		t.Errorf(
			"List() returned %d tasks, want %d",
			len(storedTasks),
			writerCount,
		)
	}
}
