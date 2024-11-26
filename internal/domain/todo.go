package domain

import (
	"errors"
	"time"
)

type TodoItem struct {
	ID          string
	Description string
	DueDate     time.Time
	FileID      string
}
type TodoItemCreateEvent struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	FileID      string    `json:"file_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (t *TodoItem) Validate() error {
	if t.Description == "" {
		return errors.New("description cannot be empty")
	}
	if time.Now().After(t.DueDate) {
		return errors.New("due date must be in the future")
	}
	return nil
}
