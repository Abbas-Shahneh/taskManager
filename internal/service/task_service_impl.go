package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
	"github.com/Abbas-Shahneh/taskManager/internal/repository"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
	MaxOffset       = 1_000_000
)

func (s *taskService) Create(
	ctx context.Context,
	input CreateTaskInput,
) (domain.Task, error) {
	now := time.Now().UTC()

	status := input.Status

	if status == "" {
		status = domain.TaskStatusPending
	}

	task := domain.Task{
		ID:          uuid.New(),
		Title:       strings.TrimSpace(input.Title),
		Description: normalizeOptionalString(input.Description),
		Status:      status,
		Assignee:    normalizeOptionalString(input.Assignee),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := task.Validate(); err != nil {
		return domain.Task{}, fmt.Errorf(
			"validate task: %w",
			err,
		)
	}

	createdTask, err := s.repository.Create(ctx, task)
	if err != nil {
		return domain.Task{}, fmt.Errorf(
			"create task: %w",
			err,
		)
	}

	s.invalidateTaskListCache(ctx)

	return createdTask, nil
}

func (s *taskService) invalidateTaskListCache(
	ctx context.Context,
) {
	if s.cache == nil {
		return
	}

	_ = s.cache.InvalidateTaskLists(ctx)
}

func (s *taskService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Task, error) {
	if id == uuid.Nil {
		return domain.Task{}, ErrInvalidTaskID
	}

	task, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf(
			"get task: %w",
			err,
		)
	}

	return task, nil
}

func (s *taskService) List(
	ctx context.Context,
	params ListTasksParams,
) (ListTasksResult, error) {
	page := params.Page

	if page == 0 {
		page = DefaultPage
	}

	pageSize := params.PageSize

	if pageSize == 0 {
		pageSize = DefaultPageSize
	}

	if page < 1 {
		return ListTasksResult{}, ErrInvalidPage
	}

	if pageSize < 1 || pageSize > MaxPageSize {
		return ListTasksResult{}, ErrInvalidPageSize
	}

	offset := (page - 1) * pageSize

	if offset > MaxOffset {
		return ListTasksResult{}, ErrInvalidPage
	}

	params.Page = page
	params.PageSize = pageSize

	if s.cache != nil {
		if result, ok := s.getCachedTaskList(ctx, params); ok {
			return result, nil
		}
	}

	repositoryParams := repository.TaskListParams{
		Status:   params.Status,
		Assignee: params.Assignee,
		Limit:    pageSize,
		Offset:   offset,
	}

	tasks, total, err := s.repository.List(
		ctx,
		repositoryParams,
	)
	if err != nil {
		return ListTasksResult{}, fmt.Errorf(
			"list tasks: %w",
			err,
		)
	}

	totalPages := 0

	if total > 0 {
		totalPages = int(
			math.Ceil(float64(total) / float64(pageSize)),
		)
	}

	result := ListTasksResult{
		Tasks:      tasks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	if s.cache != nil {
		s.setCachedTaskList(ctx, params, result)
	}

	return result, nil
}

func (s *taskService) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateTaskInput,
) (domain.Task, error) {
	if id == uuid.Nil {
		return domain.Task{}, ErrInvalidTaskID
	}

	task := domain.Task{
		ID:          id,
		Title:       strings.TrimSpace(input.Title),
		Description: normalizeOptionalString(input.Description),
		Status:      input.Status,
		Assignee:    normalizeOptionalString(input.Assignee),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := task.Validate(); err != nil {
		return domain.Task{}, fmt.Errorf(
			"validate task: %w",
			err,
		)
	}

	updatedTask, err := s.repository.Update(
		ctx,
		task,
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf(
			"update task: %w",
			err,
		)
	}

	s.invalidateTaskListCache(ctx)

	return updatedTask, nil
}

func (s *taskService) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	if id == uuid.Nil {
		return ErrInvalidTaskID
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf(
			"delete task: %w",
			err,
		)
	}

	s.invalidateTaskListCache(ctx)

	return nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.TrimSpace(*value)

	if normalized == "" {
		return nil
	}

	return &normalized
}

func (s *taskService) getCachedTaskList(
	ctx context.Context,
	params ListTasksParams,
) (ListTasksResult, bool) {
	key := taskListCacheKey(params)

	value, err := s.cache.Get(ctx, key)
	if err != nil {
		// Cache failure or miss must never break the request.
		return ListTasksResult{}, false
	}

	var result ListTasksResult

	if err := json.Unmarshal(value, &result); err != nil {
		// Corrupt cache data is treated as a cache miss.
		return ListTasksResult{}, false
	}

	return result, true
}

func (s *taskService) setCachedTaskList(
	ctx context.Context,
	params ListTasksParams,
	result ListTasksResult,
) {
	value, err := json.Marshal(result)
	if err != nil {
		return
	}

	_ = s.cache.Set(
		ctx,
		taskListCacheKey(params),
		value,
		s.cacheTTL,
	)
}
