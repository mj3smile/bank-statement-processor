package upload

import (
	"time"
)

type (
	ID     string
	Status string
)

const (
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusProcessing Status = "processing"

	MessageProcessing string = "CSV is still being processed"
)

type Task struct {
	ID          ID
	Status      Status
	Filename    string
	Message     string
	StartedAt   time.Time
	CompletedAt time.Time
}
