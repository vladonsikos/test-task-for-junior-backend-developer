package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)

	SetRecurrence(ctx context.Context, s *taskdomain.RecurrenceSettings) (*taskdomain.RecurrenceSettings, error)
	GetRecurrence(ctx context.Context, taskID int64) (*taskdomain.RecurrenceSettings, error)
	DeleteRecurrence(ctx context.Context, taskID int64) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	GetOccurrences(ctx context.Context, taskID int64, from, to time.Time) ([]time.Time, error)
}

type RecurrenceSettingsInput struct {
	Type       taskdomain.RecurrenceType
	Interval   *int
	DayOfMonth *int
	Dates      []time.Time
	Parity     *taskdomain.Parity
}

type CreateInput struct {
	Title              string
	Description        string
	Status             taskdomain.Status
	RecurrenceSettings *RecurrenceSettingsInput
}

type UpdateInput struct {
	Title              string
	Description        string
	Status             taskdomain.Status
	RecurrenceSettings *RecurrenceSettingsInput
	RemoveRecurrence   bool
}
