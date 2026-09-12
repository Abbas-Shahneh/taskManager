package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Abbas-Shahneh/taskManager/internal/cache"
	"github.com/Abbas-Shahneh/taskManager/internal/domain"
	"github.com/Abbas-Shahneh/taskManager/internal/repository"
)

type TaskService interface {
	Create(
		ctx context.Context,
		input CreateTaskInput,
	) (domain.Task, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (domain.Task, error)

	List(
		ctx context.Context,
		params ListTasksParams,
	) (ListTasksResult, error)

	Update(
		ctx context.Context,
		id uuid.UUID,
		input UpdateTaskInput,
	) (domain.Task, error)

	Delete(
		ctx context.Context,
		id uuid.UUID,
	) error

	Count(
		context.Context,
	) (int, error)
}

type CreateTaskInput struct {
	Title       string
	Description *string
	Status      domain.TaskStatus
	Assignee    *string
}

type UpdateTaskInput struct {
	Title       string
	Description *string
	Status      domain.TaskStatus
	Assignee    *string
}

type ListTasksParams struct {
	Status   *domain.TaskStatus
	Assignee *string

	Page     int
	PageSize int
}

type ListTasksResult struct {
	Tasks      []domain.Task
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type taskService struct {
	repository repository.TaskRepository
	cache      cache.TaskListCache
	cacheTTL   time.Duration
}

func NewTaskService(
	repository repository.TaskRepository,
) TaskService {
	return &taskService{
		repository: repository,
	}
}

func NewTaskServiceWithCache(
	repository repository.TaskRepository,
	taskCache cache.TaskListCache,
	cacheTTL time.Duration,
) TaskService {
	return &taskService{
		repository: repository,
		cache:      taskCache,
		cacheTTL:   cacheTTL,
	}
}

func (s *taskService) Count(ctx context.Context) (int, error) {
	count, err := s.repository.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count tasks: %w", err)
	}

	return count, nil
}
