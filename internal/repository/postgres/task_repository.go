package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, status, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt)
	return scanTask(row)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, status, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) SetRecurrence(ctx context.Context, s *taskdomain.RecurrenceSettings) (*taskdomain.RecurrenceSettings, error) {
	const query = `
		INSERT INTO recurrence_settings (task_id, type, interval, day_of_month, dates, parity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (task_id) DO UPDATE SET
			type         = EXCLUDED.type,
			interval     = EXCLUDED.interval,
			day_of_month = EXCLUDED.day_of_month,
			dates        = EXCLUDED.dates,
			parity       = EXCLUDED.parity,
			updated_at   = EXCLUDED.updated_at
		RETURNING id, task_id, type, interval, day_of_month, dates, parity, created_at, updated_at
	`

	var parity *string
	if s.Parity != nil {
		p := string(*s.Parity)
		parity = &p
	}

	row := r.pool.QueryRow(ctx, query,
		s.TaskID, string(s.Type), s.Interval, s.DayOfMonth,
		s.Dates, parity, s.CreatedAt, s.UpdatedAt,
	)

	return scanRecurrence(row)
}

func (r *Repository) GetRecurrence(ctx context.Context, taskID int64) (*taskdomain.RecurrenceSettings, error) {
	const query = `
		SELECT id, task_id, type, interval, day_of_month, dates, parity, created_at, updated_at
		FROM recurrence_settings
		WHERE task_id = $1
	`

	row := r.pool.QueryRow(ctx, query, taskID)
	rs, err := scanRecurrence(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return rs, nil
}

func (r *Repository) DeleteRecurrence(ctx context.Context, taskID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM recurrence_settings WHERE task_id = $1`, taskID)
	return err
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}

func scanRecurrence(scanner taskScanner) (*taskdomain.RecurrenceSettings, error) {
	var (
		rs        taskdomain.RecurrenceSettings
		recType   string
		parityStr *string
		dates     []time.Time
	)

	if err := scanner.Scan(
		&rs.ID,
		&rs.TaskID,
		&recType,
		&rs.Interval,
		&rs.DayOfMonth,
		&dates,
		&parityStr,
		&rs.CreatedAt,
		&rs.UpdatedAt,
	); err != nil {
		return nil, err
	}

	rs.Type = taskdomain.RecurrenceType(recType)

	if parityStr != nil {
		p := taskdomain.Parity(*parityStr)
		rs.Parity = &p
	}

	for _, d := range dates {
		rs.Dates = append(rs.Dates, d.UTC())
	}

	return &rs, nil
}
