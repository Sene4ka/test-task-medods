package handlers

import (
	taskdomain "example.com/taskservice/internal/domain/task"
	"time"
)

type recurrenceDTO struct {
	Type       taskdomain.RecurrenceType `json:"type"`
	Interval   int                       `json:"interval,omitempty"`
	DayOfMonth int                       `json:"day_of_month,omitempty"`
	Dates      []string                  `json:"dates,omitempty"`
	Parity     taskdomain.Parity         `json:"parity,omitempty"`
	StartDate  string                    `json:"start_date"`
	EndDate    string                    `json:"end_date,omitempty"`
}

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  *recurrenceDTO    `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  *recurrenceDTO    `json:"recurrence,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type occurrenceDTO struct {
	Date string  `json:"date"`
	Task taskDTO `json:"task"`
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
	if task.Recurrence != nil {
		r := task.Recurrence
		dto.Recurrence = &recurrenceDTO{
			Type:       r.Type,
			Interval:   r.Interval,
			DayOfMonth: r.DayOfMonth,
			Dates:      r.Dates,
			Parity:     r.Parity,
			StartDate:  r.StartDate,
			EndDate:    r.EndDate,
		}
	}
	return dto
}

func dtoToRecurrence(dto *recurrenceDTO) *taskdomain.Recurrence {
	if dto == nil {
		return nil
	}
	return &taskdomain.Recurrence{
		Type:       dto.Type,
		Interval:   dto.Interval,
		DayOfMonth: dto.DayOfMonth,
		Dates:      dto.Dates,
		Parity:     dto.Parity,
		StartDate:  dto.StartDate,
		EndDate:    dto.EndDate,
	}
}
