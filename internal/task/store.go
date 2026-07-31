package task

import "context"

// Store defines the persistence operations required by Service.
type Store interface {
	List(ctx context.Context) ([]Task, error)
	Create(ctx context.Context, task Task) (Task, error)
	Update(ctx context.Context, task Task) (Task, error)
	Delete(ctx context.Context, id uint64) error
}
