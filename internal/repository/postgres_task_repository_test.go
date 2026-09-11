package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/require"
)

func TestPostgresTaskRepository_Create(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	taskID := uuid.New()
	now := time.Now().UTC()

	description := "test description"
	assignee := "alice"

	task := domain.Task{
		ID:          taskID,
		Title:       "Test task",
		Description: &description,
		Status:      domain.TaskStatusPending,
		Assignee:    &assignee,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := regexp.QuoteMeta(`
		INSERT INTO tasks (
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
	`)

	db.ExpectQuery(query).
		WithArgs(
			task.ID,
			task.Title,
			task.Description,
			task.Status,
			task.Assignee,
			task.CreatedAt,
			task.UpdatedAt,
		).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee",
				"created_at",
				"updated_at",
			}).
				AddRow(
					task.ID,
					task.Title,
					task.Description,
					task.Status,
					task.Assignee,
					task.CreatedAt,
					task.UpdatedAt,
				),
		)

	result, err := repo.Create(context.Background(), task)

	require.NoError(t, err)
	require.Equal(t, task, result)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_Create_Error(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	task := testTask("Create error")

	query := regexp.QuoteMeta(`
		INSERT INTO tasks (
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
	`)

	db.ExpectQuery(query).
		WithArgs(
			task.ID,
			task.Title,
			task.Description,
			task.Status,
			task.Assignee,
			task.CreatedAt,
			task.UpdatedAt,
		).
		WillReturnError(errors.New("database error"))

	_, err = repo.Create(context.Background(), task)

	require.Error(t, err)
	require.ErrorContains(t, err, "create task")
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_GetByID(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	taskID := uuid.New()
	now := time.Now().UTC()

	task := domain.Task{
		ID:        taskID,
		Title:     "Find me",
		Status:    domain.TaskStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := regexp.QuoteMeta(`
		SELECT
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1
	`)

	db.ExpectQuery(query).
		WithArgs(taskID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee",
				"created_at",
				"updated_at",
			}).
				AddRow(
					task.ID,
					task.Title,
					task.Description,
					task.Status,
					task.Assignee,
					task.CreatedAt,
					task.UpdatedAt,
				),
		)

	result, err := repo.GetByID(context.Background(), taskID)

	require.NoError(t, err)
	require.Equal(t, task, result)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_GetByID_NotFound(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	taskID := uuid.New()

	query := regexp.QuoteMeta(`
		SELECT
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1
	`)

	db.ExpectQuery(query).
		WithArgs(taskID).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(context.Background(), taskID)

	require.Error(t, err)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_GetByID_DatabaseError(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	taskID := uuid.New()

	query := regexp.QuoteMeta(`
		SELECT
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1
	`)

	db.ExpectQuery(query).
		WithArgs(taskID).
		WillReturnError(errors.New("database error"))

	_, err = repo.GetByID(context.Background(), taskID)

	require.Error(t, err)
	require.ErrorContains(t, err, "get task")
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_Count(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	query := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM tasks
	`)

	db.ExpectQuery(query).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(7),
		)

	count, err := repo.Count(context.Background())

	require.NoError(t, err)
	require.Equal(t, 7, count)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_Count_Error(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	query := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM tasks
	`)

	db.ExpectQuery(query).
		WillReturnError(errors.New("database error"))

	_, err = repo.Count(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "count tasks")
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_List(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	status := domain.TaskStatusInProgress
	assignee := "alice"

	taskID := uuid.New()
	now := time.Now().UTC()

	description := "test description"

	task := domain.Task{
		ID:          taskID,
		Title:       "Test task",
		Description: &description,
		Status:      status,
		Assignee:    &assignee,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	params := TaskListParams{
		Status:   &status,
		Assignee: &assignee,
		Limit:    20,
		Offset:   0,
	}

	countQuery := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM tasks
		WHERE status = $1 AND assignee = $2
	`)

	db.ExpectQuery(countQuery).
		WithArgs(status, assignee).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	listQuery := regexp.QuoteMeta(`
		SELECT
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
		FROM tasks
		WHERE status = $1 AND assignee = $2
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`)

	db.ExpectQuery(listQuery).
		WithArgs(status, assignee, 20, 0).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee",
				"created_at",
				"updated_at",
			}).
				AddRow(
					task.ID,
					task.Title,
					task.Description,
					task.Status,
					task.Assignee,
					task.CreatedAt,
					task.UpdatedAt,
				),
		)

	tasks, total, err := repo.List(context.Background(), params)

	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, tasks, 1)
	require.Equal(t, task, tasks[0])
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_List_CountError(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	params := TaskListParams{
		Limit:  20,
		Offset: 0,
	}

	countQuery := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM tasks
	`)

	db.ExpectQuery(countQuery).
		WillReturnError(errors.New("database error"))

	_, _, err = repo.List(context.Background(), params)

	require.Error(t, err)
	require.ErrorContains(t, err, "count tasks")
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_List_QueryError(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	params := TaskListParams{
		Limit:  20,
		Offset: 0,
	}

	countQuery := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM tasks
	`)

	db.ExpectQuery(countQuery).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(0),
		)

	listQuery := regexp.QuoteMeta(`
		SELECT
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
		FROM tasks

		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`)

	db.ExpectQuery(listQuery).
		WithArgs(20, 0).
		WillReturnError(errors.New("database error"))

	_, _, err = repo.List(context.Background(), params)

	require.Error(t, err)
	require.ErrorContains(t, err, "list tasks")
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_List_ScanError(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	params := TaskListParams{
		Limit:  20,
		Offset: 0,
	}

	countQuery := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM tasks
	`)

	db.ExpectQuery(countQuery).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	listQuery := regexp.QuoteMeta(`
		SELECT
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
		FROM tasks

		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`)

	db.ExpectQuery(listQuery).
		WithArgs(20, 0).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee",
				"created_at",
				"updated_at",
			}).
				AddRow(
					"not-a-uuid",
					"Test task",
					nil,
					domain.TaskStatusPending,
					nil,
					time.Now().UTC(),
					time.Now().UTC(),
				),
		)

	_, _, err = repo.List(context.Background(), params)

	require.Error(t, err)
	require.ErrorContains(t, err, "scan task")
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_List_RowsError(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	params := TaskListParams{
		Limit:  20,
		Offset: 0,
	}

	countQuery := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM tasks
	`)

	db.ExpectQuery(countQuery).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(0),
		)

	listQuery := regexp.QuoteMeta(`
		SELECT
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
		FROM tasks

		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`)

	rows := pgxmock.NewRows([]string{
		"id",
		"title",
		"description",
		"status",
		"assignee",
		"created_at",
		"updated_at",
	}).RowError(0, errors.New("row iteration error"))

	db.ExpectQuery(listQuery).
		WithArgs(20, 0).
		WillReturnRows(rows)

	_, _, err = repo.List(context.Background(), params)

	require.Error(t, err)
	require.ErrorContains(t, err, "iterate tasks")
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_UpdateUnit(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	task := testTask("Updated task")
	task.Description = stringPtr("updated description")
	task.Assignee = stringPtr("bob")
	task.Status = domain.TaskStatusCompleted

	query := regexp.QuoteMeta(`
		UPDATE tasks
		SET
			title = $2,
			description = $3,
			status = $4,
			assignee = $5,
			updated_at = $6
		WHERE id = $1
		RETURNING
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
	`)

	db.ExpectQuery(query).
		WithArgs(
			task.ID,
			task.Title,
			task.Description,
			task.Status,
			task.Assignee,
			task.UpdatedAt,
		).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee",
				"created_at",
				"updated_at",
			}).
				AddRow(
					task.ID,
					task.Title,
					task.Description,
					task.Status,
					task.Assignee,
					task.CreatedAt,
					task.UpdatedAt,
				),
		)

	result, err := repo.Update(context.Background(), task)

	require.NoError(t, err)
	require.Equal(t, task, result)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_Update_NotFound(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	task := testTask("Not found")

	query := regexp.QuoteMeta(`
		UPDATE tasks
		SET
			title = $2,
			description = $3,
			status = $4,
			assignee = $5,
			updated_at = $6
		WHERE id = $1
		RETURNING
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
	`)

	db.ExpectQuery(query).
		WithArgs(
			task.ID,
			task.Title,
			task.Description,
			task.Status,
			task.Assignee,
			task.UpdatedAt,
		).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.Update(context.Background(), task)

	require.ErrorIs(t, err, domain.ErrTaskNotFound)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_Update_DatabaseError(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	task := testTask("Update error")

	query := regexp.QuoteMeta(`
		UPDATE tasks
		SET
			title = $2,
			description = $3,
			status = $4,
			assignee = $5,
			updated_at = $6
		WHERE id = $1
		RETURNING
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
	`)

	db.ExpectQuery(query).
		WithArgs(
			task.ID,
			task.Title,
			task.Description,
			task.Status,
			task.Assignee,
			task.UpdatedAt,
		).
		WillReturnError(errors.New("database error"))

	_, err = repo.Update(context.Background(), task)

	require.Error(t, err)
	require.ErrorContains(t, err, "update task")
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_DeleteUnit(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	taskID := uuid.New()

	query := regexp.QuoteMeta(`
		DELETE FROM tasks
		WHERE id = $1
	`)

	db.ExpectExec(query).
		WithArgs(taskID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(context.Background(), taskID)

	require.NoError(t, err)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_Delete_NotFound(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	taskID := uuid.New()

	query := regexp.QuoteMeta(`
		DELETE FROM tasks
		WHERE id = $1
	`)

	db.ExpectExec(query).
		WithArgs(taskID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err = repo.Delete(context.Background(), taskID)

	require.ErrorIs(t, err, domain.ErrTaskNotFound)
	require.NoError(t, db.ExpectationsWereMet())
}

func TestPostgresTaskRepository_Delete_DatabaseError(t *testing.T) {
	db, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer db.Close()

	repo := &PostgresTaskRepository{
		db: db,
	}

	taskID := uuid.New()

	query := regexp.QuoteMeta(`
		DELETE FROM tasks
		WHERE id = $1
	`)

	db.ExpectExec(query).
		WithArgs(taskID).
		WillReturnError(errors.New("database error"))

	err = repo.Delete(context.Background(), taskID)

	require.Error(t, err)
	require.ErrorContains(t, err, "delete task")
	require.NoError(t, db.ExpectationsWereMet())
}

func stringPtr(value string) *string {
	return &value
}
