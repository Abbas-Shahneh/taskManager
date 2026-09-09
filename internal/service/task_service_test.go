package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
	"github.com/Abbas-Shahneh/taskManager/internal/repository"
)

func TestTaskService_Create(t *testing.T) {
	t.Run("creates task with default status", func(t *testing.T) {
		var received domain.Task

		repo := &repository.MockTaskRepository{
			CreateFunc: func(
				_ context.Context,
				task domain.Task,
			) (domain.Task, error) {
				received = task
				return task, nil
			},
		}

		service := NewTaskService(repo)

		task, err := service.Create(
			context.Background(),
			CreateTaskInput{
				Title: "  Buy milk  ",
			},
		)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if task.ID == uuid.Nil {
			t.Fatal("expected generated task ID")
		}

		if task.Title != "Buy milk" {
			t.Fatalf(
				"expected normalized title, got %q",
				task.Title,
			)
		}

		if task.Status != domain.TaskStatusPending {
			t.Fatalf(
				"expected pending status, got %q",
				task.Status,
			)
		}

		if received.ID != task.ID {
			t.Fatal("repository received different task ID")
		}
	})

	t.Run("creates task with supplied fields", func(t *testing.T) {
		description := "Get two bottles"
		assignee := "john"

		repo := &repository.MockTaskRepository{
			CreateFunc: func(
				_ context.Context,
				task domain.Task,
			) (domain.Task, error) {
				return task, nil
			},
		}

		service := NewTaskService(repo)

		task, err := service.Create(
			context.Background(),
			CreateTaskInput{
				Title:       "Buy milk",
				Description: &description,
				Status:      domain.TaskStatusInProgress,
				Assignee:    &assignee,
			},
		)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if task.Description == nil ||
			*task.Description != description {
			t.Fatalf("unexpected description")
		}

		if task.Assignee == nil ||
			*task.Assignee != assignee {
			t.Fatalf("unexpected assignee")
		}

		if task.Status != domain.TaskStatusInProgress {
			t.Fatalf("unexpected status")
		}
	})

	t.Run("rejects invalid title", func(t *testing.T) {
		called := false

		repo := &repository.MockTaskRepository{
			CreateFunc: func(
				_ context.Context,
				task domain.Task,
			) (domain.Task, error) {
				called = true
				return task, nil
			},
		}

		service := NewTaskService(repo)

		_, err := service.Create(
			context.Background(),
			CreateTaskInput{
				Title: "   ",
			},
		)

		if !errors.Is(err, domain.ErrInvalidTaskTitle) {
			t.Fatalf(
				"expected ErrInvalidTaskTitle, got %v",
				err,
			)
		}

		if called {
			t.Fatal("repository should not be called")
		}
	})

	t.Run("rejects invalid status", func(t *testing.T) {
		repo := &repository.MockTaskRepository{
			CreateFunc: func(
				_ context.Context,
				task domain.Task,
			) (domain.Task, error) {
				return task, nil
			},
		}

		service := NewTaskService(repo)

		_, err := service.Create(
			context.Background(),
			CreateTaskInput{
				Title:  "Test",
				Status: "invalid",
			},
		)

		if !errors.Is(err, domain.ErrInvalidTaskStatus) {
			t.Fatalf(
				"expected ErrInvalidTaskStatus, got %v",
				err,
			)
		}
	})

	t.Run("propagates repository failure", func(t *testing.T) {
		expectedErr := errors.New("database unavailable")

		repo := &repository.MockTaskRepository{
			CreateFunc: func(
				_ context.Context,
				task domain.Task,
			) (domain.Task, error) {
				return domain.Task{}, expectedErr
			},
		}

		service := NewTaskService(repo)

		_, err := service.Create(
			context.Background(),
			CreateTaskInput{
				Title: "Test",
			},
		)

		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"expected wrapped repository error, got %v",
				err,
			)
		}
	})
}

func TestTaskService_GetByID(t *testing.T) {
	id := uuid.New()

	expectedTask := domain.Task{
		ID:        id,
		Title:     "Test",
		Status:    domain.TaskStatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	t.Run("returns task", func(t *testing.T) {
		repo := &repository.MockTaskRepository{
			GetByIDFunc: func(
				_ context.Context,
				gotID uuid.UUID,
			) (domain.Task, error) {
				if gotID != id {
					t.Fatalf("unexpected ID: %s", gotID)
				}

				return expectedTask, nil
			},
		}

		service := NewTaskService(repo)

		task, err := service.GetByID(
			context.Background(),
			id,
		)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if task.ID != id {
			t.Fatalf("unexpected task ID")
		}
	})

	t.Run("rejects nil UUID", func(t *testing.T) {
		called := false

		repo := &repository.MockTaskRepository{
			GetByIDFunc: func(
				_ context.Context,
				id uuid.UUID,
			) (domain.Task, error) {
				called = true
				return expectedTask, nil
			},
		}

		service := NewTaskService(repo)

		_, err := service.GetByID(
			context.Background(),
			uuid.Nil,
		)

		if !errors.Is(err, ErrInvalidTaskID) {
			t.Fatalf(
				"expected ErrInvalidTaskID, got %v",
				err,
			)
		}

		if called {
			t.Fatal("repository should not be called")
		}
	})

	t.Run("propagates not found", func(t *testing.T) {
		repo := &repository.MockTaskRepository{
			GetByIDFunc: func(
				_ context.Context,
				id uuid.UUID,
			) (domain.Task, error) {
				return domain.Task{}, domain.ErrTaskNotFound
			},
		}

		service := NewTaskService(repo)

		_, err := service.GetByID(
			context.Background(),
			id,
		)

		if !errors.Is(err, domain.ErrTaskNotFound) {
			t.Fatalf(
				"expected ErrTaskNotFound, got %v",
				err,
			)
		}
	})
}

func TestTaskService_List(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		repo := &repository.MockTaskRepository{
			ListFunc: func(
				_ context.Context,
				params repository.TaskListParams,
			) ([]domain.Task, int, error) {
				if params.Limit != DefaultPageSize {
					t.Fatalf(
						"expected limit %d, got %d",
						DefaultPageSize,
						params.Limit,
					)
				}

				if params.Offset != 0 {
					t.Fatalf(
						"expected offset 0, got %d",
						params.Offset,
					)
				}

				return []domain.Task{
					{
						ID:     uuid.New(),
						Title:  "Test",
						Status: domain.TaskStatusPending,
					},
				}, 1, nil
			},
		}

		service := NewTaskService(repo)

		result, err := service.List(
			context.Background(),
			ListTasksParams{},
		)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if result.Page != 1 {
			t.Fatalf("expected page 1, got %d", result.Page)
		}

		if result.PageSize != 20 {
			t.Fatalf(
				"expected page size 20, got %d",
				result.PageSize,
			)
		}

		if result.TotalPages != 1 {
			t.Fatalf(
				"expected total pages 1, got %d",
				result.TotalPages,
			)
		}
	})

	t.Run("calculates pagination", func(t *testing.T) {
		repo := &repository.MockTaskRepository{
			ListFunc: func(
				_ context.Context,
				params repository.TaskListParams,
			) ([]domain.Task, int, error) {
				if params.Limit != 20 {
					t.Fatalf("unexpected limit")
				}

				if params.Offset != 40 {
					t.Fatalf(
						"expected offset 40, got %d",
						params.Offset,
					)
				}

				return nil, 55, nil
			},
		}

		service := NewTaskService(repo)

		result, err := service.List(
			context.Background(),
			ListTasksParams{
				Page:     3,
				PageSize: 20,
			},
		)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if result.Total != 55 {
			t.Fatalf("unexpected total")
		}

		if result.TotalPages != 3 {
			t.Fatalf(
				"expected 3 pages, got %d",
				result.TotalPages,
			)
		}
	})

	t.Run("rejects invalid page", func(t *testing.T) {
		repo := &repository.MockTaskRepository{}

		service := NewTaskService(repo)

		_, err := service.List(
			context.Background(),
			ListTasksParams{
				Page: -1,
			},
		)

		if !errors.Is(err, ErrInvalidPage) {
			t.Fatalf(
				"expected ErrInvalidPage, got %v",
				err,
			)
		}
	})

	t.Run("rejects page size above maximum", func(t *testing.T) {
		repo := &repository.MockTaskRepository{}

		service := NewTaskService(repo)

		_, err := service.List(
			context.Background(),
			ListTasksParams{
				Page:     1,
				PageSize: MaxPageSize + 1,
			},
		)

		if !errors.Is(err, ErrInvalidPageSize) {
			t.Fatalf(
				"expected ErrInvalidPageSize, got %v",
				err,
			)
		}
	})

	t.Run("propagates repository failure", func(t *testing.T) {
		expectedErr := errors.New("database failure")

		repo := &repository.MockTaskRepository{
			ListFunc: func(
				_ context.Context,
				params repository.TaskListParams,
			) ([]domain.Task, int, error) {
				return nil, 0, expectedErr
			},
		}

		service := NewTaskService(repo)

		_, err := service.List(
			context.Background(),
			ListTasksParams{},
		)

		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"expected repository error, got %v",
				err,
			)
		}
	})
}

func TestTaskService_Update(t *testing.T) {
	id := uuid.New()

	t.Run("updates task", func(t *testing.T) {
		repo := &repository.MockTaskRepository{
			UpdateFunc: func(
				_ context.Context,
				task domain.Task,
			) (domain.Task, error) {
				if task.ID != id {
					t.Fatalf("unexpected task ID")
				}

				return task, nil
			},
		}

		service := NewTaskService(repo)

		task, err := service.Update(
			context.Background(),
			id,
			UpdateTaskInput{
				Title:  "Updated task",
				Status: domain.TaskStatusCompleted,
			},
		)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if task.Title != "Updated task" {
			t.Fatalf("unexpected title")
		}

		if task.Status != domain.TaskStatusCompleted {
			t.Fatalf("unexpected status")
		}

		if task.UpdatedAt.IsZero() {
			t.Fatal("expected UpdatedAt")
		}
	})

	t.Run("rejects invalid ID", func(t *testing.T) {
		called := false

		repo := &repository.MockTaskRepository{
			UpdateFunc: func(
				_ context.Context,
				task domain.Task,
			) (domain.Task, error) {
				called = true
				return task, nil
			},
		}

		service := NewTaskService(repo)

		_, err := service.Update(
			context.Background(),
			uuid.Nil,
			UpdateTaskInput{
				Title:  "Test",
				Status: domain.TaskStatusPending,
			},
		)

		if !errors.Is(err, ErrInvalidTaskID) {
			t.Fatalf(
				"expected ErrInvalidTaskID, got %v",
				err,
			)
		}

		if called {
			t.Fatal("repository should not be called")
		}
	})

	t.Run("rejects invalid input", func(t *testing.T) {
		repo := &repository.MockTaskRepository{}

		service := NewTaskService(repo)

		_, err := service.Update(
			context.Background(),
			id,
			UpdateTaskInput{
				Title:  "",
				Status: domain.TaskStatusPending,
			},
		)

		if !errors.Is(err, domain.ErrInvalidTaskTitle) {
			t.Fatalf(
				"expected ErrInvalidTaskTitle, got %v",
				err,
			)
		}
	})

	t.Run("propagates repository failure", func(t *testing.T) {
		expectedErr := errors.New("database failure")

		repo := &repository.MockTaskRepository{
			UpdateFunc: func(
				_ context.Context,
				task domain.Task,
			) (domain.Task, error) {
				return domain.Task{}, expectedErr
			},
		}

		service := NewTaskService(repo)

		_, err := service.Update(
			context.Background(),
			id,
			UpdateTaskInput{
				Title:  "Test",
				Status: domain.TaskStatusPending,
			},
		)

		if !errors.Is(err, expectedErr) {
			t.Fatalf(
				"expected repository error, got %v",
				err,
			)
		}
	})
}

func TestTaskService_Delete(t *testing.T) {
	id := uuid.New()

	t.Run("deletes task", func(t *testing.T) {
		called := false

		repo := &repository.MockTaskRepository{
			DeleteFunc: func(
				_ context.Context,
				gotID uuid.UUID,
			) error {
				called = true

				if gotID != id {
					t.Fatalf("unexpected ID")
				}

				return nil
			},
		}

		service := NewTaskService(repo)

		if err := service.Delete(
			context.Background(),
			id,
		); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		if !called {
			t.Fatal("expected repository Delete to be called")
		}
	})

	t.Run("rejects nil ID", func(t *testing.T) {
		called := false

		repo := &repository.MockTaskRepository{
			DeleteFunc: func(
				_ context.Context,
				id uuid.UUID,
			) error {
				called = true
				return nil
			},
		}

		service := NewTaskService(repo)

		err := service.Delete(
			context.Background(),
			uuid.Nil,
		)

		if !errors.Is(err, ErrInvalidTaskID) {
			t.Fatalf(
				"expected ErrInvalidTaskID, got %v",
				err,
			)
		}

		if called {
			t.Fatal("repository should not be called")
		}
	})

	t.Run("propagates not found", func(t *testing.T) {
		repo := &repository.MockTaskRepository{
			DeleteFunc: func(
				_ context.Context,
				id uuid.UUID,
			) error {
				return domain.ErrTaskNotFound
			},
		}

		service := NewTaskService(repo)

		err := service.Delete(
			context.Background(),
			id,
		)

		if !errors.Is(err, domain.ErrTaskNotFound) {
			t.Fatalf(
				"expected ErrTaskNotFound, got %v",
				err,
			)
		}
	})
}
