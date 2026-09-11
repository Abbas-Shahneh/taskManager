package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
)

type PostgresTaskRepository struct {
	db *pgxpool.Pool
}

func NewPostgresTaskRepository(
	db *pgxpool.Pool,
) *PostgresTaskRepository {
	return &PostgresTaskRepository{
		db: db,
	}
}

func (r *PostgresTaskRepository) Create(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {
	const query = `
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
	`

	err := r.db.QueryRow(
		ctx,
		query,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.Assignee,
		task.CreatedAt,
		task.UpdatedAt,
	).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Assignee,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		return domain.Task{}, fmt.Errorf(
			"create task: %w",
			err,
		)
	}

	return task, nil
}

func (r *PostgresTaskRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Task, error) {
	const query = `
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
	`

	var task domain.Task

	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Assignee,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Task{}, domain.ErrTaskNotFound
		}

		return domain.Task{}, fmt.Errorf(
			"get task: %w",
			err,
		)
	}

	return task, nil
}

func (r *PostgresTaskRepository) List(
	ctx context.Context,
	params TaskListParams,
) ([]domain.Task, int, error) {
	var (
		where    []string
		args     []any
		argIndex = 1
	)

	if params.Status != nil {
		where = append(
			where,
			fmt.Sprintf("status = $%d", argIndex),
		)

		args = append(args, *params.Status)
		argIndex++
	}

	if params.Assignee != nil {
		where = append(
			where,
			fmt.Sprintf("assignee = $%d", argIndex),
		)

		args = append(args, *params.Assignee)
		argIndex++
	}

	whereClause := ""

	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM tasks
		%s
	`, whereClause)

	var total int

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf(
			"count tasks: %w",
			err,
		)
	}

	listQuery := fmt.Sprintf(`
		SELECT
			id,
			title,
			description,
			status,
			assignee,
			created_at,
			updated_at
		FROM tasks
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.Query(
		ctx,
		listQuery,
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf(
			"list tasks: %w",
			err,
		)
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)

	for rows.Next() {
		var task domain.Task

		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Assignee,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf(
				"scan task: %w",
				err,
			)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf(
			"iterate tasks: %w",
			err,
		)
	}

	return tasks, total, nil
}

func (r *PostgresTaskRepository) Count(
	ctx context.Context,
) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM tasks
	`

	var count int

	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count tasks: %w", err)
	}

	return count, nil
}

func (r *PostgresTaskRepository) Update(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {
	const query = `
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
	`

	err := r.db.QueryRow(
		ctx,
		query,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.Assignee,
		task.UpdatedAt,
	).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Assignee,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Task{}, domain.ErrTaskNotFound
		}

		return domain.Task{}, fmt.Errorf(
			"update task: %w",
			err,
		)
	}

	return task, nil
}

func (r *PostgresTaskRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	const query = `
		DELETE FROM tasks
		WHERE id = $1
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf(
			"delete task: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}
