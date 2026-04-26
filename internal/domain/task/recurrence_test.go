package task_test

import (
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func date(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	if err != nil {
		panic(err)
	}
	return t
}

func dates(ss ...string) []time.Time {
	out := make([]time.Time, len(ss))
	for i, s := range ss {
		out[i] = date(s)
	}
	return out
}

func occurrenceDates(occ []taskdomain.Occurrence) []time.Time {
	out := make([]time.Time, len(occ))
	for i, o := range occ {
		out[i] = o.Date
	}
	return out
}

func equalDates(a, b []time.Time) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
}

func TestOccurrences_Daily(t *testing.T) {
	tests := []struct {
		name     string
		rec      taskdomain.Recurrence
		from     string
		to       string
		expected []string
	}{
		{
			name: "every day, full week",
			rec: taskdomain.Recurrence{
				Type:      taskdomain.RecurrenceDaily,
				Interval:  1,
				StartDate: "2026-05-01",
			},
			from:     "2026-05-01",
			to:       "2026-05-03",
			expected: []string{"2026-05-01", "2026-05-02", "2026-05-03"},
		},
		{
			name: "every 2 days",
			rec: taskdomain.Recurrence{
				Type:      taskdomain.RecurrenceDaily,
				Interval:  2,
				StartDate: "2026-05-01",
			},
			from:     "2026-05-01",
			to:       "2026-05-07",
			expected: []string{"2026-05-01", "2026-05-03", "2026-05-05", "2026-05-07"},
		},
		{
			name: "every 7 days",
			rec: taskdomain.Recurrence{
				Type:      taskdomain.RecurrenceDaily,
				Interval:  7,
				StartDate: "2026-05-01",
			},
			from:     "2026-05-01",
			to:       "2026-05-31",
			expected: []string{"2026-05-01", "2026-05-08", "2026-05-15", "2026-05-22", "2026-05-29"},
		},
		{
			name: "interval defaults to 1 when zero",
			rec: taskdomain.Recurrence{
				Type:      taskdomain.RecurrenceDaily,
				Interval:  0,
				StartDate: "2026-05-01",
			},
			from:     "2026-05-01",
			to:       "2026-05-03",
			expected: []string{"2026-05-01", "2026-05-02", "2026-05-03"},
		},
		{
			name: "from is before start_date — clamps to start_date",
			rec: taskdomain.Recurrence{
				Type:      taskdomain.RecurrenceDaily,
				Interval:  1,
				StartDate: "2026-05-03",
			},
			from:     "2026-05-01",
			to:       "2026-05-05",
			expected: []string{"2026-05-03", "2026-05-04", "2026-05-05"},
		},
		{
			name: "end_date cuts off results",
			rec: taskdomain.Recurrence{
				Type:      taskdomain.RecurrenceDaily,
				Interval:  1,
				StartDate: "2026-05-01",
				EndDate:   "2026-05-03",
			},
			from:     "2026-05-01",
			to:       "2026-05-10",
			expected: []string{"2026-05-01", "2026-05-02", "2026-05-03"},
		},
		{
			name:     "range does not intersect start_date",
			rec:      taskdomain.Recurrence{Type: taskdomain.RecurrenceDaily, Interval: 1, StartDate: "2026-06-01"},
			from:     "2026-05-01",
			to:       "2026-05-31",
			expected: nil,
		},
		{
			name:     "from equals to — single day match",
			rec:      taskdomain.Recurrence{Type: taskdomain.RecurrenceDaily, Interval: 1, StartDate: "2026-05-01"},
			from:     "2026-05-05",
			to:       "2026-05-05",
			expected: []string{"2026-05-05"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &taskdomain.Task{Recurrence: &tt.rec}
			occ, err := task.Occurrences(date(tt.from), date(tt.to))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := occurrenceDates(occ)
			want := dates(tt.expected...)
			if !equalDates(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

func TestOccurrences_Monthly(t *testing.T) {
	tests := []struct {
		name       string
		dayOfMonth int
		startDate  string
		endDate    string
		from       string
		to         string
		expected   []string
	}{
		{
			name:       "15th of each month",
			dayOfMonth: 15,
			startDate:  "2026-01-01",
			from:       "2026-01-01",
			to:         "2026-04-30",
			expected:   []string{"2026-01-15", "2026-02-15", "2026-03-15", "2026-04-15"},
		},
		{
			name:       "31st — skips months without day 31",
			dayOfMonth: 31,
			startDate:  "2026-01-01",
			from:       "2026-01-01",
			to:         "2026-08-31",
			expected:   []string{"2026-01-31", "2026-03-31", "2026-05-31", "2026-07-31", "2026-08-31"},
		},
		{
			name:       "30th — skips february",
			dayOfMonth: 30,
			startDate:  "2026-01-01",
			from:       "2026-01-01",
			to:         "2026-05-31",
			expected:   []string{"2026-01-30", "2026-03-30", "2026-04-30", "2026-05-30"},
		},
		{
			name:       "29th in non-leap year — skips february",
			dayOfMonth: 29,
			startDate:  "2026-01-01",
			from:       "2026-01-01",
			to:         "2026-04-30",
			expected:   []string{"2026-01-29", "2026-03-29", "2026-04-29"},
		},
		{
			name:       "29th in leap year — includes february",
			dayOfMonth: 29,
			startDate:  "2024-01-01",
			from:       "2024-01-01",
			to:         "2024-04-30",
			expected:   []string{"2024-01-29", "2024-02-29", "2024-03-29", "2024-04-29"},
		},
		{
			name:       "28th — present in every month",
			dayOfMonth: 28,
			startDate:  "2026-01-01",
			from:       "2026-01-01",
			to:         "2026-04-30",
			expected:   []string{"2026-01-28", "2026-02-28", "2026-03-28", "2026-04-28"},
		},
		{
			name:       "from is after the day in the first month — starts next month",
			dayOfMonth: 10,
			startDate:  "2026-01-01",
			from:       "2026-01-15",
			to:         "2026-03-31",
			expected:   []string{"2026-02-10", "2026-03-10"},
		},
		{
			name:       "end_date cuts off",
			dayOfMonth: 15,
			startDate:  "2026-01-01",
			endDate:    "2026-02-20",
			from:       "2026-01-01",
			to:         "2026-04-30",
			expected:   []string{"2026-01-15", "2026-02-15"},
		},
		{
			name:       "no matches in range",
			dayOfMonth: 31,
			startDate:  "2026-01-01",
			from:       "2026-04-01",
			to:         "2026-04-30",
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := taskdomain.Recurrence{
				Type:       taskdomain.RecurrenceMonthly,
				DayOfMonth: tt.dayOfMonth,
				StartDate:  tt.startDate,
				EndDate:    tt.endDate,
			}
			task := &taskdomain.Task{Recurrence: &rec}
			occ, err := task.Occurrences(date(tt.from), date(tt.to))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := occurrenceDates(occ)
			want := dates(tt.expected...)
			if !equalDates(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

func TestOccurrences_SpecificDates(t *testing.T) {
	tests := []struct {
		name     string
		dates    []string
		from     string
		to       string
		expected []string
	}{
		{
			name:     "all dates within range",
			dates:    []string{"2026-06-01", "2026-09-01", "2026-12-01"},
			from:     "2026-01-01",
			to:       "2026-12-31",
			expected: []string{"2026-06-01", "2026-09-01", "2026-12-01"},
		},
		{
			name:     "some dates outside range",
			dates:    []string{"2026-06-01", "2026-09-01", "2026-12-01"},
			from:     "2026-06-01",
			to:       "2026-10-01",
			expected: []string{"2026-06-01", "2026-09-01"},
		},
		{
			name:     "all dates outside range",
			dates:    []string{"2026-06-01", "2026-09-01"},
			from:     "2026-01-01",
			to:       "2026-05-31",
			expected: nil,
		},
		{
			name:     "single date — exact boundary match",
			dates:    []string{"2026-05-15"},
			from:     "2026-05-15",
			to:       "2026-05-15",
			expected: []string{"2026-05-15"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := taskdomain.Recurrence{
				Type:      taskdomain.RecurrenceSpecificDates,
				Dates:     tt.dates,
				StartDate: tt.from,
			}
			task := &taskdomain.Task{Recurrence: &rec}
			occ, err := task.Occurrences(date(tt.from), date(tt.to))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := occurrenceDates(occ)
			want := dates(tt.expected...)
			if !equalDates(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

func TestOccurrences_EvenOdd(t *testing.T) {
	tests := []struct {
		name     string
		parity   taskdomain.Parity
		from     string
		to       string
		expected []string
	}{
		{
			name:     "even days in may 1–10",
			parity:   taskdomain.ParityEven,
			from:     "2026-05-01",
			to:       "2026-05-10",
			expected: []string{"2026-05-02", "2026-05-04", "2026-05-06", "2026-05-08", "2026-05-10"},
		},
		{
			name:     "odd days in may 1–10",
			parity:   taskdomain.ParityOdd,
			from:     "2026-05-01",
			to:       "2026-05-10",
			expected: []string{"2026-05-01", "2026-05-03", "2026-05-05", "2026-05-07", "2026-05-09"},
		},
		{
			name:     "month boundary — 31 is odd, 1 is odd",
			parity:   taskdomain.ParityOdd,
			from:     "2026-05-30",
			to:       "2026-06-02",
			expected: []string{"2026-05-31", "2026-06-01"},
		},
		{
			name:     "single even day — match",
			parity:   taskdomain.ParityEven,
			from:     "2026-05-04",
			to:       "2026-05-04",
			expected: []string{"2026-05-04"},
		},
		{
			name:     "single even day — no match",
			parity:   taskdomain.ParityEven,
			from:     "2026-05-05",
			to:       "2026-05-05",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := taskdomain.Recurrence{
				Type:      taskdomain.RecurrenceEvenOdd,
				Parity:    tt.parity,
				StartDate: tt.from,
			}
			task := &taskdomain.Task{Recurrence: &rec}
			occ, err := task.Occurrences(date(tt.from), date(tt.to))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := occurrenceDates(occ)
			want := dates(tt.expected...)
			if !equalDates(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

func TestOccurrences_NoRecurrence(t *testing.T) {
	task := &taskdomain.Task{}
	occ, err := task.Occurrences(date("2026-05-01"), date("2026-05-31"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(occ) != 0 {
		t.Errorf("expected no occurrences for task without recurrence, got %d", len(occ))
	}
}
