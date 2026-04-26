package task

import (
	"context"
	"fmt"
	"sort"
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
		Recurrence:  normalized.Recurrence,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Create(ctx, model)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}
	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:  normalized.Recurrence,
		UpdatedAt:   s.now(),
	}
	return s.repo.Update(ctx, model)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

// ListOccurrences returns all occurrences of recurring tasks within the given date range.
func (s *Service) ListOccurrences(ctx context.Context, fromStr, toStr string) ([]taskdomain.Occurrence, error) {
	from, err := time.ParseInLocation("2006-01-02", fromStr, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid 'from' date: %v", ErrInvalidInput, err)
	}
	to, err := time.ParseInLocation("2006-01-02", toStr, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid 'to' date: %v", ErrInvalidInput, err)
	}
	if to.Before(from) {
		return nil, fmt.Errorf("%w: 'to' must not be before 'from'", ErrInvalidInput)
	}
	if to.Sub(from) > 366*24*time.Hour {
		return nil, fmt.Errorf("%w: date range must not exceed 366 days", ErrInvalidInput)
	}

	tasks, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var all []taskdomain.Occurrence
	for i := range tasks {
		if tasks[i].Recurrence == nil {
			continue
		}
		occ, err := tasks[i].Occurrences(from, to)
		if err != nil {
			return nil, err
		}
		all = append(all, occ...)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Date.Before(all[j].Date)
	})

	return all, nil
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
	if err := validateRecurrence(input.Recurrence); err != nil {
		return CreateInput{}, err
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
	if err := validateRecurrence(input.Recurrence); err != nil {
		return UpdateInput{}, err
	}
	return input, nil
}

func validateRecurrence(r *taskdomain.Recurrence) error {
	if r == nil {
		return nil
	}
	if r.StartDate == "" {
		return fmt.Errorf("%w: recurrence.start_date is required", ErrInvalidInput)
	}
	if _, err := time.Parse("2006-01-02", r.StartDate); err != nil {
		return fmt.Errorf("%w: recurrence.start_date must be YYYY-MM-DD", ErrInvalidInput)
	}
	if r.EndDate != "" {
		if _, err := time.Parse("2006-01-02", r.EndDate); err != nil {
			return fmt.Errorf("%w: recurrence.end_date must be YYYY-MM-DD", ErrInvalidInput)
		}
	}

	switch r.Type {
	case taskdomain.RecurrenceDaily:
		if r.Interval < 0 {
			return fmt.Errorf("%w: recurrence.interval must be >= 0", ErrInvalidInput)
		}
	case taskdomain.RecurrenceMonthly:
		if r.DayOfMonth < 1 || r.DayOfMonth > 31 {
			return fmt.Errorf("%w: recurrence.day_of_month must be between 1 and 31", ErrInvalidInput)
		}
	case taskdomain.RecurrenceSpecificDates:
		if len(r.Dates) == 0 {
			return fmt.Errorf("%w: recurrence.dates must not be empty", ErrInvalidInput)
		}
		for _, d := range r.Dates {
			if _, err := time.Parse("2006-01-02", d); err != nil {
				return fmt.Errorf("%w: recurrence.dates contains invalid date %q", ErrInvalidInput, d)
			}
		}
	case taskdomain.RecurrenceEvenOdd:
		if r.Parity != taskdomain.ParityEven && r.Parity != taskdomain.ParityOdd {
			return fmt.Errorf("%w: recurrence.parity must be 'even' or 'odd'", ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: unknown recurrence type %q", ErrInvalidInput, r.Type)
	}
	return nil
}
