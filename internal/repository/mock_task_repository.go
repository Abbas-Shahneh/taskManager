package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
)

type MockTaskRepository struct {
	CreateFunc  func(context.Context, domain.Task) (domain.Task, error)
	GetByIDFunc func(context.Context, uuid.UUID) (domain.Task, error)
	ListFunc    func(context.Context, TaskListParams) ([]domain.Task, int, error)
	UpdateFunc  func(context.Context, domain.Task) (domain.Task, error)
	DeleteFunc  func(context.Context, uuid.UUID) error
}

func (m *MockTaskRepository) Create(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {
	return m.CreateFunc(ctx, task)
}

func (m *MockTaskRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Task, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockTaskRepository) List(
	ctx context.Context,
	params TaskListParams,
) ([]domain.Task, int, error) {
	return m.ListFunc(ctx, params)
}

func (m *MockTaskRepository) Update(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {
	return m.UpdateFunc(ctx, task)
}

func (m *MockTaskRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	return m.DeleteFunc(ctx, id)
}
