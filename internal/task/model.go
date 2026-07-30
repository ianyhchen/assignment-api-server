package task

type Status int

const (
	StatusIncomplete Status = 0
	StatusCompleted  Status = 1
)

func (status Status) IsValid() bool {
	switch status {
	case StatusCompleted, StatusIncomplete:
		return true
	default:
		return false
	}
}

type Task struct {
	ID     uint64
	Name   string
	Status Status
}
