package task

type Status int

const (
	StatusIncomplete Status = 0
	StatusCompleted  Status = 1
)

type Task struct {
	ID     uint64
	Name   string
	Status Status
}
