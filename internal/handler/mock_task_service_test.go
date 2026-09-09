package handler

import (
	"context"

	"github.com/google/uuid"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
	"github.com/Abbas-Shahneh/taskManager/internal/service"
)

type mockTaskService struct {
	CreateFunc func(
		context.Context,
		service.CreateTaskInput,
	) (domain.Task, error)

	GetByIDFunc func(
		context.Context,
		uuid.UUID,
	) (domain.Task, error)

	ListFunc func(
		context.Context,
		service.ListTasksParams,
	) (service.ListTasksResult, error)

	UpdateFunc func(
		context.Context,
		uuid.UUID,
		service.UpdateTaskInput,
	) (domain.Task, error)

	DeleteFunc func(
		context.Context,
		uuid.UUID,
	) error
}

func (m *mockTaskService) Create(
	ctx context.Context,
	input service.CreateTaskInput,
) (domain.Task, error) {
	return m.CreateFunc(ctx, input)
}

func (m *mockTaskService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Task, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *mockTaskService) List(
	ctx context.Context,
	params service.ListTasksParams,
) (service.ListTasksResult, error) {
	return m.ListFunc(ctx, params)
}

func (m *mockTaskService) Update(
	ctx context.Context,
	id uuid.UUID,
	input service.UpdateTaskInput,
) (domain.Task, error) {
	return m.UpdateFunc(ctx, id, input)
}

func (m *mockTaskService) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	return m.DeleteFunc(ctx, id)
}
