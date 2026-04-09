package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	if input.RecurrenceSettings != nil {
		rs := recurrenceInputToModel(input.RecurrenceSettings, created.ID, now)
		if err := rs.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
		}
		saved, err := s.repo.SetRecurrence(ctx, rs)
		if err != nil {
			return nil, err
		}
		created.RecurrenceSettings = saved
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rs, err := s.repo.GetRecurrence(ctx, id)
	if err != nil {
		return nil, err
	}
	task.RecurrenceSettings = rs

	return task, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   now,
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	if input.RemoveRecurrence {
		if err := s.repo.DeleteRecurrence(ctx, id); err != nil {
			return nil, err
		}
	} else if input.RecurrenceSettings != nil {
		rs := recurrenceInputToModel(input.RecurrenceSettings, id, now)
		if err := rs.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
		}
		saved, err := s.repo.SetRecurrence(ctx, rs)
		if err != nil {
			return nil, err
		}
		updated.RecurrenceSettings = saved
	} else {
		rs, err := s.repo.GetRecurrence(ctx, id)
		if err != nil {
			return nil, err
		}
		updated.RecurrenceSettings = rs
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	tasks, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range tasks {
		rs, err := s.repo.GetRecurrence(ctx, tasks[i].ID)
		if err != nil {
			return nil, err
		}
		tasks[i].RecurrenceSettings = rs
	}

	return tasks, nil
}

func (s *Service) GetOccurrences(ctx context.Context, taskID int64, from, to time.Time) ([]time.Time, error) {
	if taskID <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	if from.After(to) {
		return nil, fmt.Errorf("%w: from must be before or equal to to", ErrInvalidInput)
	}
	if to.Sub(from).Hours() > 24*366 {
		return nil, fmt.Errorf("%w: date range must not exceed 1 year", ErrInvalidInput)
	}

	rs, err := s.repo.GetRecurrence(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if rs == nil {
		return nil, fmt.Errorf("%w: task has no recurrence settings", ErrInvalidInput)
	}

	return rs.GenerateOccurrences(from, to), nil
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}
	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	return input, nil
}

func recurrenceInputToModel(input *RecurrenceSettingsInput, taskID int64, now time.Time) *taskdomain.RecurrenceSettings {
	return &taskdomain.RecurrenceSettings{
		TaskID:     taskID,
		Type:       input.Type,
		Interval:   input.Interval,
		DayOfMonth: input.DayOfMonth,
		Dates:      input.Dates,
		Parity:     input.Parity,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
