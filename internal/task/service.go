package task

import (
	"context"
	"fmt"
	"strings"
)

// CreateInput contains the client-editable fields required to create a task.
type CreateInput struct {
	Name   string
	Status Status
}

// UpdateInput contains the ID and replacement fields required to update a task.
type UpdateInput struct {
	ID     uint64
	Name   string
	Status Status
}

// Service applies task rules and coordinates persistence operations.
type Service struct {
	store Store
}

// NewService creates a task service backed by store.
func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

// Create validates, normalizes, and stores a new task.
func (service *Service) Create(ctx context.Context, input CreateInput) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	normalizedName, err := normalizeTaskFields(input.Name, input.Status)
	if err != nil {
		return Task{}, err
	}

	newTask := Task{
		Name:   normalizedName,
		Status: input.Status,
	}

	createdTask, err := service.store.Create(ctx, newTask)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return createdTask, nil
}

// normalizeTaskFields applies the validation shared by create and update.
func normalizeTaskFields(name string, status Status) (string, error) {
	normalizedName := strings.TrimSpace(name)

	if normalizedName == "" {
		return "", ErrInvalidName
	}

	if !status.IsValid() {
		return "", ErrInvalidStatus
	}

	return normalizedName, nil
}

// List returns all tasks from the configured store.
func (service *Service) List(ctx context.Context) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	tasks, err := service.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	return tasks, nil
}

// Update validates and replaces an existing task.
func (service *Service) Update(ctx context.Context, input UpdateInput) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	if input.ID == 0 {
		return Task{}, ErrInvalidTaskID
	}

	normalizedName, err := normalizeTaskFields(input.Name, input.Status)
	if err != nil {
		return Task{}, err
	}

	updatedTask := Task{
		ID:     input.ID,
		Name:   normalizedName,
		Status: input.Status,
	}

	result, err := service.store.Update(ctx, updatedTask)
	if err != nil {
		return Task{}, fmt.Errorf("update task %d: %w", input.ID, err)
	}

	return result, nil
}

// Delete removes an existing task by ID.
func (service *Service) Delete(ctx context.Context, id uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if id == 0 {
		return ErrInvalidTaskID
	}

	if err := service.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete task %d: %w", id, err)
	}

	return nil
}
