package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrInvalidTaskStatus  = errors.New("invalid task status")
	ErrInvalidTaskTitle   = errors.New("invalid task title")
	ErrInvalidDescription = errors.New("invalid task description")
	ErrInvalidAssignee    = errors.New("invalid task assignee")
)

type Task struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	Assignee    *string    `json:"assignee,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (t Task) Validate() error {
	t.Title = strings.TrimSpace(t.Title)

	if t.Title == "" || len(t.Title) > 200 {
		return ErrInvalidTaskTitle
	}

	if t.Description != nil {
		description := strings.TrimSpace(*t.Description)

		if len(description) > 5000 {
			return ErrInvalidDescription
		}
	}

	if !IsValidTaskStatus(t.Status) {
		return ErrInvalidTaskStatus
	}

	if t.Assignee != nil {
		assignee := strings.TrimSpace(*t.Assignee)

		if len(assignee) > 100 {
			return ErrInvalidAssignee
		}
	}

	return nil
}

func IsValidTaskStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusPending,
		TaskStatusInProgress,
		TaskStatusCompleted:
		return true
	default:
		return false
	}
}
