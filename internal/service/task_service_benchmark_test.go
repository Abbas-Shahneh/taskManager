package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
	"github.com/Abbas-Shahneh/taskManager/internal/repository"
)

func BenchmarkTaskService_Create(b *testing.B) {
	repo := &repository.MockTaskRepository{
		CreateFunc: func(ctx context.Context, task domain.Task) (domain.Task, error) {
			return task, nil
		},
	}

	svc := NewTaskService(repo)

	input := CreateTaskInput{
		Title:  "benchmark task",
		Status: domain.TaskStatusPending,
	}

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := svc.Create(ctx, input); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTaskService_GetByID(b *testing.B) {
	taskID := uuid.New()

	repo := &repository.MockTaskRepository{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.Task, error) {
			return domain.Task{
				ID:     id,
				Title:  "benchmark task",
				Status: domain.TaskStatusPending,
			}, nil
		},
	}

	svc := NewTaskService(repo)

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := svc.GetByID(ctx, taskID); err != nil {
			b.Fatal(err)
		}
	}
}
