package task

import "errors"

var (
	ErrTaskNotFound  = errors.New("task not found")
	ErrInvalidTaskID = errors.New("invalid task ID")
	ErrInvalidName   = errors.New("invalid task name")
	ErrInvalidStatus = errors.New("invalid task status")
)
