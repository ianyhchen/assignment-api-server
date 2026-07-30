package task

import (
	"context"
	"fmt"
	"strings"
)

type CreateInput struct {
	Name   string
	Status Status
}

type UpdateInput struct {
	ID     uint64
	Name   string
	Status Status
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

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
