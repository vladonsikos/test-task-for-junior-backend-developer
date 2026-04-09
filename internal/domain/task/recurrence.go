package task

import (
    "errors"
    "time"
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

// RecurrenceSettings описывает настройки периодичности задачи.
// В зависимости от Type заполняются разные поля:
//   - daily:          Interval (каждые N дней, >= 1)
//   - monthly:        DayOfMonth (число месяца, 1-30)
//   - specific_dates: Dates (список конкретных дат)
//   - even_odd:       Parity (even или odd)
type RecurrenceSettings struct {
    ID          int64
    TaskID      int64
    Type        RecurrenceType
    Interval    *int       // для daily
    DayOfMonth  *int       // для monthly
    Dates       []time.Time // для specific_dates
    Parity      *Parity    // для even_odd
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

var (
    ErrInvalidRecurrenceType     = errors.New("invalid recurrence type")
    ErrInvalidRecurrenceInterval = errors.New("interval must be >= 1")
    ErrInvalidDayOfMonth         = errors.New("day_of_month must be between 1 and 30")
    ErrInvalidParity             = errors.New("parity must be 'even' or 'odd'")
    ErrNoDates                   = errors.New("specific_dates requires at least one date")
)

func (r RecurrenceType) Valid() bool {
    switch r {
    case RecurrenceDaily, RecurrenceMonthly, RecurrenceSpecificDates, RecurrenceEvenOdd:
        return true
    default:
        return false
    }
}

func (p Parity) Valid() bool {
    return p == ParityEven || p == ParityOdd
}

// Validate проверяет корректность настроек в зависимости от типа.
func (r *RecurrenceSettings) Validate() error {
    if !r.Type.Valid() {
        return ErrInvalidRecurrenceType
    }

    switch r.Type {
    case RecurrenceDaily:
        if r.Interval == nil || *r.Interval < 1 {
            return ErrInvalidRecurrenceInterval
        }
    case RecurrenceMonthly:
        if r.DayOfMonth == nil || *r.DayOfMonth < 1 || *r.DayOfMonth > 30 {
            return ErrInvalidDayOfMonth
        }
    case RecurrenceSpecificDates:
        if len(r.Dates) == 0 {
            return ErrNoDates
        }
    case RecurrenceEvenOdd:
        if r.Parity == nil || !r.Parity.Valid() {
            return ErrInvalidParity
        }
    }

    return nil
}

// GenerateOccurrences возвращает все даты вхождений в диапазоне [from, to].
// Для daily отсчёт ведётся от from (первый день диапазона).
func (r *RecurrenceSettings) GenerateOccurrences(from, to time.Time) []time.Time {
    var result []time.Time

    from = from.UTC().Truncate(24 * time.Hour)
    to = to.UTC().Truncate(24 * time.Hour)

    for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
        if r.occursOn(d, from) {
            result = append(result, d)
        }
    }

    return result
}

func (r *RecurrenceSettings) occursOn(date, origin time.Time) bool {
    date = date.UTC().Truncate(24 * time.Hour)
    day := date.Day()

    switch r.Type {
    case RecurrenceDaily:
        diff := int(date.Sub(origin).Hours() / 24)
        if diff < 0 {
            return false
        }
        return diff%*r.Interval == 0
    case RecurrenceMonthly:
        return day == *r.DayOfMonth
    case RecurrenceSpecificDates:
        for _, d := range r.Dates {
            if d.UTC().Truncate(24 * time.Hour).Equal(date) {
                return true
            }
        }
        return false
    case RecurrenceEvenOdd:
        if *r.Parity == ParityEven {
            return day%2 == 0
        }
        return day%2 != 0
    }
    return false
}
