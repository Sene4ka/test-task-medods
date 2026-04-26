package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEvenOdd       RecurrenceType = "even_odd"
)

type Parity string

const (
	ParityEven Parity = "even"
	ParityOdd  Parity = "odd"
)

// Recurrence describes the repetition rule for a task.
type Recurrence struct {
	Type RecurrenceType `json:"type"`

	// daily: repeat every N days (defaults to 1)
	Interval int `json:"interval,omitempty"`

	// monthly: day of the month to repeat on (1–31)
	DayOfMonth int `json:"day_of_month,omitempty"`

	// specific_dates: explicit list of dates in "YYYY-MM-DD" format
	Dates []string `json:"dates,omitempty"`

	// even_odd: "even" or "odd" calendar days of the month
	Parity Parity `json:"parity,omitempty"`

	// Rule active range
	StartDate string `json:"start_date"`         // "YYYY-MM-DD", required
	EndDate   string `json:"end_date,omitempty"` // "YYYY-MM-DD", optional
}

type Task struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      Status      `json:"status"`
	Recurrence  *Recurrence `json:"recurrence,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
