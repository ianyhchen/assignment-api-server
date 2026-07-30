package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/ianyhchen/assignment-api-server/internal/task"
)

type TaskStore struct {
	mu     sync.RWMutex
	tasks  map[uint64]task.Task
	nextID uint64
}

// Ensure at compile time that *TaskStore implements task.Store.
var _ task.Store = (*TaskStore)(nil)

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks: make(map[uint64]task.Task),
	}
}

func (store *TaskStore) Create(ctx context.Context, newTask task.Task) (task.Task, error) {
	if err := ctx.Err(); err != nil {
		return task.Task{}, err
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	store.nextID++
	newTask.ID = store.nextID
	store.tasks[newTask.ID] = newTask

	return newTask, nil
}

func (store *TaskStore) List(ctx context.Context) ([]task.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	store.mu.RLock()

	tasks := make([]task.Task, 0, len(store.tasks))
	for _, storedTask := range store.tasks {
		tasks = append(tasks, storedTask)
	}

	store.mu.RUnlock()

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})

	return tasks, nil
}

func (store *TaskStore) Update(ctx context.Context, updatedTask task.Task) (task.Task, error) {
	if err := ctx.Err(); err != nil {
		return task.Task{}, err
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.tasks[updatedTask.ID]; !exists {
		return task.Task{}, task.ErrTaskNotFound
	}

	store.tasks[updatedTask.ID] = updatedTask

	return updatedTask, nil
}

func (store *TaskStore) Delete(ctx context.Context, id uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.tasks[id]; !exists {
		return task.ErrTaskNotFound
	}

	delete(store.tasks, id)

	return nil
}
