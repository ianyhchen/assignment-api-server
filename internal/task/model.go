package task

// Status represents whether a task is incomplete or completed.
type Status int

const (
	// StatusIncomplete identifies a task that has not been completed.
	StatusIncomplete Status = 0
	// StatusCompleted identifies a completed task.
	StatusCompleted Status = 1
)

// IsValid reports whether the status is supported by the task domain.
func (status Status) IsValid() bool {
	switch status {
	case StatusCompleted, StatusIncomplete:
		return true
	default:
		return false
	}
}

// Task represents a task managed by the application.
type Task struct {
	ID     uint64
	Name   string
	Status Status
}
