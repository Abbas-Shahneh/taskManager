package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
)

func setupPostgresIntegrationTest(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)

	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(ctx))

	_, err = pool.Exec(ctx, `
		TRUNCATE TABLE tasks
	`)
	require.NoError(t, err)

	return pool, ctx
}

func testTask(title string) domain.Task {
	now := time.Now().UTC()

	description := "integration test description"
	assignee := "integration-user"

	return domain.Task{
		ID:          uuid.New(),
		Title:       title,
		Description: &description,
		Status:      domain.TaskStatusPending,
		Assignee:    &assignee,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestPostgresTaskRepository_CreateAndGetByID(t *testing.T) {
	pool, ctx := setupPostgresIntegrationTest(t)
	repo := NewPostgresTaskRepository(pool)

	task := testTask("Create integration test")

	created, err := repo.Create(ctx, task)
	require.NoError(t, err)

	require.Equal(t, task.ID, created.ID)
	require.Equal(t, task.Title, created.Title)
	require.Equal(t, task.Description, created.Description)
	require.Equal(t, task.Status, created.Status)
	require.Equal(t, task.Assignee, created.Assignee)
	require.False(t, created.CreatedAt.IsZero())
	require.False(t, created.UpdatedAt.IsZero())

	found, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	require.Equal(t, created.ID, found.ID)
	require.Equal(t, created.Title, found.Title)
	require.Equal(t, created.Description, found.Description)
	require.Equal(t, created.Status, found.Status)
	require.Equal(t, created.Assignee, found.Assignee)
}

func TestPostgresTaskRepository_GetByIDNotFound(t *testing.T) {
	pool, ctx := setupPostgresIntegrationTest(t)
	repo := NewPostgresTaskRepository(pool)

	_, err := repo.GetByID(ctx, uuid.New())

	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrTaskNotFound))
}

func TestPostgresTaskRepository_ListAndFilters(t *testing.T) {
	pool, ctx := setupPostgresIntegrationTest(t)
	repo := NewPostgresTaskRepository(pool)

	assigneeA := "alice"
	assigneeB := "bob"

	tasks := []domain.Task{
		{
			ID:        uuid.New(),
			Title:     "Pending Alice",
			Status:    domain.TaskStatusPending,
			Assignee:  &assigneeA,
			CreatedAt: time.Now().UTC().Add(-3 * time.Minute),
			UpdatedAt: time.Now().UTC(),
		},
		{
			ID:        uuid.New(),
			Title:     "Completed Alice",
			Status:    domain.TaskStatusCompleted,
			Assignee:  &assigneeA,
			CreatedAt: time.Now().UTC().Add(-2 * time.Minute),
			UpdatedAt: time.Now().UTC(),
		},
		{
			ID:        uuid.New(),
			Title:     "Pending Bob",
			Status:    domain.TaskStatusPending,
			Assignee:  &assigneeB,
			CreatedAt: time.Now().UTC().Add(-1 * time.Minute),
			UpdatedAt: time.Now().UTC(),
		},
	}

	for _, task := range tasks {
		_, err := repo.Create(ctx, task)
		require.NoError(t, err)
	}

	allTasks, total, err := repo.List(ctx, TaskListParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)

	require.Equal(t, 3, total)
	require.Len(t, allTasks, 3)

	status := domain.TaskStatusPending

	pendingTasks, total, err := repo.List(ctx, TaskListParams{
		Status: &status,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)

	require.Equal(t, 2, total)
	require.Len(t, pendingTasks, 2)

	assignee := "alice"

	aliceTasks, total, err := repo.List(ctx, TaskListParams{
		Assignee: &assignee,
		Limit:    10,
		Offset:   0,
	})
	require.NoError(t, err)

	require.Equal(t, 2, total)
	require.Len(t, aliceTasks, 2)
}

func TestPostgresTaskRepository_ListPagination(t *testing.T) {
	pool, ctx := setupPostgresIntegrationTest(t)
	repo := NewPostgresTaskRepository(pool)

	for i := 0; i < 5; i++ {
		task := testTask("Pagination test")
		task.CreatedAt = time.Now().UTC().Add(time.Duration(-i) * time.Minute)

		_, err := repo.Create(ctx, task)
		require.NoError(t, err)
	}

	firstPage, total, err := repo.List(ctx, TaskListParams{
		Limit:  2,
		Offset: 0,
	})
	require.NoError(t, err)

	require.Equal(t, 5, total)
	require.Len(t, firstPage, 2)

	secondPage, total, err := repo.List(ctx, TaskListParams{
		Limit:  2,
		Offset: 2,
	})
	require.NoError(t, err)

	require.Equal(t, 5, total)
	require.Len(t, secondPage, 2)

	require.NotEqual(t, firstPage[0].ID, secondPage[0].ID)
	require.NotEqual(t, firstPage[1].ID, secondPage[1].ID)
}

func TestPostgresTaskRepository_Update(t *testing.T) {
	pool, ctx := setupPostgresIntegrationTest(t)
	repo := NewPostgresTaskRepository(pool)

	task := testTask("Before update")

	_, err := repo.Create(ctx, task)
	require.NoError(t, err)

	description := "updated description"
	assignee := "updated-user"

	updated := task
	updated.Title = "After update"
	updated.Description = &description
	updated.Status = domain.TaskStatusInProgress
	updated.Assignee = &assignee
	updated.UpdatedAt = time.Now().UTC()

	result, err := repo.Update(ctx, updated)
	require.NoError(t, err)

	require.Equal(t, updated.ID, result.ID)
	require.Equal(t, "After update", result.Title)
	require.Equal(t, &description, result.Description)
	require.Equal(t, domain.TaskStatusInProgress, result.Status)
	require.Equal(t, &assignee, result.Assignee)

	found, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)

	require.Equal(t, "After update", found.Title)
	require.Equal(t, domain.TaskStatusInProgress, found.Status)
}

func TestPostgresTaskRepository_UpdateNotFound(t *testing.T) {
	pool, ctx := setupPostgresIntegrationTest(t)
	repo := NewPostgresTaskRepository(pool)

	task := testTask("Missing update")

	_, err := repo.Update(ctx, task)

	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrTaskNotFound))
}

func TestPostgresTaskRepository_Delete(t *testing.T) {
	pool, ctx := setupPostgresIntegrationTest(t)
	repo := NewPostgresTaskRepository(pool)

	task := testTask("Delete test")

	_, err := repo.Create(ctx, task)
	require.NoError(t, err)

	err = repo.Delete(ctx, task.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, task.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrTaskNotFound))
}

func TestPostgresTaskRepository_DeleteNotFound(t *testing.T) {
	pool, ctx := setupPostgresIntegrationTest(t)
	repo := NewPostgresTaskRepository(pool)

	err := repo.Delete(ctx, uuid.New())

	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrTaskNotFound))
}
