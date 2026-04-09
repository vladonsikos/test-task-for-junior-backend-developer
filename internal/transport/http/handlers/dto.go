package handlers

import (
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type recurrenceSettingsInputDTO struct {
	Type       taskdomain.RecurrenceType `json:"type"`
	Interval   *int                      `json:"interval,omitempty"`
	DayOfMonth *int                      `json:"day_of_month,omitempty"`
	Dates      []string                  `json:"dates,omitempty"`
	Parity     *taskdomain.Parity        `json:"parity,omitempty"`
}

type taskMutationDTO struct {
	Title              string                      `json:"title"`
	Description        string                      `json:"description"`
	Status             taskdomain.Status           `json:"status"`
	RecurrenceSettings *recurrenceSettingsInputDTO `json:"recurrence_settings,omitempty"`
	RemoveRecurrence   bool                        `json:"remove_recurrence,omitempty"`
}

type recurrenceSettingsDTO struct {
	Type       taskdomain.RecurrenceType `json:"type"`
	Interval   *int                      `json:"interval,omitempty"`
	DayOfMonth *int                      `json:"day_of_month,omitempty"`
	Dates      []string                  `json:"dates,omitempty"`
	Parity     *taskdomain.Parity        `json:"parity,omitempty"`
}

type taskDTO struct {
	ID                 int64                  `json:"id"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description"`
	Status             taskdomain.Status      `json:"status"`
	RecurrenceSettings *recurrenceSettingsDTO `json:"recurrence_settings,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
	if task.RecurrenceSettings != nil {
		rs := task.RecurrenceSettings
		rdto := &recurrenceSettingsDTO{
			Type:       rs.Type,
			Interval:   rs.Interval,
			DayOfMonth: rs.DayOfMonth,
			Parity:     rs.Parity,
		}
		for _, d := range rs.Dates {
			rdto.Dates = append(rdto.Dates, d.UTC().Format("2006-01-02"))
		}
		dto.RecurrenceSettings = rdto
	}
	return dto
}

func recurrenceInputDTOToUsecase(dto *recurrenceSettingsInputDTO) (*taskusecase.RecurrenceSettingsInput, error) {
	if dto == nil {
		return nil, nil
	}

	input := &taskusecase.RecurrenceSettingsInput{
		Type:       dto.Type,
		Interval:   dto.Interval,
		DayOfMonth: dto.DayOfMonth,
		Parity:     dto.Parity,
	}

	for _, ds := range dto.Dates {
		t, err := time.Parse("2006-01-02", ds)
		if err != nil {
			return nil, fmt.Errorf("invalid date format %q, expected YYYY-MM-DD", ds)
		}
		input.Dates = append(input.Dates, t.UTC())
	}

	return input, nil
}
