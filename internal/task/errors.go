package task

import "errors"

var ErrTaskNotFound = errors.New("task not found")
var ErrInvalidName = errors.New("invalid task name")
var ErrInvalidStatus = errors.New("invalid task status")
