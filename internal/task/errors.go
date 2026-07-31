// Package task contains the task domain model and application service.
package task

import "errors"

var (
	// ErrTaskNotFound indicates that a requested task does not exist.
	ErrTaskNotFound = errors.New("task not found")
	// ErrInvalidTaskID indicates that a task ID is not positive.
	ErrInvalidTaskID = errors.New("invalid task ID")
	// ErrInvalidName indicates that a task name is empty after normalization.
	ErrInvalidName = errors.New("invalid task name")
	// ErrInvalidStatus indicates that a task status is outside the supported values.
	ErrInvalidStatus = errors.New("invalid task status")
)
