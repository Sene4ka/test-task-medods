package task

import (
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

// Occurrence represents a single instance of a task on a specific date.
type Occurrence struct {
	Task Task
	Date time.Time
}

// Occurrences returns all dates the task falls on within [from, to] inclusive.
// Returns nil for tasks without a recurrence rule.
func (t *Task) Occurrences(from, to time.Time) ([]Occurrence, error) {
	if t.Recurrence == nil {
		return nil, nil
	}
	r := t.Recurrence

	start, err := parseDate(r.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid recurrence start_date: %w", err)
	}

	if start.After(from) {
		from = start
	}

	if r.EndDate != "" {
		end, err := parseDate(r.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid recurrence end_date: %w", err)
		}
		if end.Before(to) {
			to = end
		}
	}

	if from.After(to) {
		return nil, nil
	}

	var dates []time.Time
	switch r.Type {
	case RecurrenceDaily:
		dates = dailyDates(from, to, r.Interval)
	case RecurrenceMonthly:
		dates = monthlyDates(from, to, r.DayOfMonth)
	case RecurrenceSpecificDates:
		dates, err = specificDates(from, to, r.Dates)
		if err != nil {
			return nil, err
		}
	case RecurrenceEvenOdd:
		dates = evenOddDates(from, to, r.Parity)
	default:
		return nil, fmt.Errorf("unknown recurrence type: %s", r.Type)
	}

	occurrences := make([]Occurrence, 0, len(dates))
	for _, d := range dates {
		occurrences = append(occurrences, Occurrence{Task: *t, Date: d})
	}
	return occurrences, nil
}

func dailyDates(from, to time.Time, interval int) []time.Time {
	if interval <= 0 {
		interval = 1
	}
	var dates []time.Time
	for d := from; !d.After(to); d = d.AddDate(0, 0, interval) {
		dates = append(dates, d)
	}
	return dates
}

func monthlyDates(from, to time.Time, dayOfMonth int) []time.Time {
	if dayOfMonth < 1 || dayOfMonth > 31 {
		return nil
	}
	var dates []time.Time

	y, m := from.Year(), from.Month()

	for {
		if time.Date(y, m, 1, 0, 0, 0, 0, time.UTC).After(to) {
			break
		}

		if dayOfMonth <= daysInMonth(y, m) {
			cur := time.Date(y, m, dayOfMonth, 0, 0, 0, 0, time.UTC)
			if !cur.Before(from) && !cur.After(to) {
				dates = append(dates, cur)
			}
		}

		m++
		if m > 12 {
			m = 1
			y++
		}
	}
	return dates
}

func daysInMonth(y int, m time.Month) int {
	return time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func specificDates(from, to time.Time, rawDates []string) ([]time.Time, error) {
	var dates []time.Time
	for _, raw := range rawDates {
		d, err := parseDate(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid date %q: %w", raw, err)
		}
		if !d.Before(from) && !d.After(to) {
			dates = append(dates, d)
		}
	}
	return dates, nil
}

func evenOddDates(from, to time.Time, parity Parity) []time.Time {
	var dates []time.Time
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		dayNum := d.Day()
		isEven := dayNum%2 == 0
		if (parity == ParityEven && isEven) || (parity == ParityOdd && !isEven) {
			dates = append(dates, d)
		}
	}
	return dates
}

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation(dateLayout, s, time.UTC)
}
