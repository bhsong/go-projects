package task

import (
	"time"
)

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"Title"`
	Done      bool      `json:"Done"`
	CreatedAt time.Time `json:"CreatedAt"`
}
