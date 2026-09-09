package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
)

type TaskListParams struct {
	Status   *domain.TaskStatus
	Assignee *string

	Limit  int
	Offset int
}

type TaskRepository interface {
	Create(
		ctx context.Context,
		task domain.Task,
	) (domain.Task, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (domain.Task, error)

	List(
		ctx context.Context,
		params TaskListParams,
	) ([]domain.Task, int, error)

	Update(
		ctx context.Context,
		task domain.Task,
	) (domain.Task, error)

	Delete(
		ctx context.Context,
		id uuid.UUID,
	) error
}
