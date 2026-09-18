package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrUnsupportedRecurrence is returned by ParseRecurrence for a rule this
// package does not implement, and for a rule that is not an RRULE at all.
//
// It is deliberately loud. The alternative — accepting the string and expanding
// the part we understand — produces a habit whose streak is plausible and wrong,
// and nothing on screen or in the data says so. A rule that cannot be expanded
// exactly is refused at the point it is written, not approximated at the point
// it is read.
var ErrUnsupportedRecurrence = errors.New("domain: unsupported recurrence rule")

// ErrWindowTooLarge is returned when an expansion window is wider than
// maxWindowDays. Expansion walks the window day by day, so the window is what
// bounds it; a caller asking for four centuries has a bug, not a habit.
var ErrWindowTooLarge = errors.New("domain: recurrence window too large")

const (
	// maxWindowDays caps Occurrences at roughly two hundred years.
	maxWindowDays = 200 * 366
	// maxSearchDays caps the unbounded search in NextOccurrence and
	// PreviousOccurrence at roughly a century. A supported rule that has not
	// produced an occurrence within a century of the starting point either has
	// none — FREQ=MONTHLY;BYMONTHDAY=31;INTERVAL=12 anchored in February never
	// matches — or is far enough away not to be a habit.
	maxSearchDays = 100 * 366
)

// Frequency is the FREQ of a supported recurrence rule.
type Frequency string

// The three supported frequencies. YEARLY, HOURLY, MINUTELY and SECONDLY are
// rejected at parse time: habits are date-granular and a yearly habit has no
// streak worth the name.
const (
	FreqDaily   Frequency = "DAILY"
	FreqWeekly  Frequency = "WEEKLY"
	FreqMonthly Frequency = "MONTHLY"
)

// Recurrence is a parsed recurrence rule, restricted to the subset this project
// supports.
//
// # Why a hand-rolled subset rather than a library
//
// There is no maintained pure-Go RRULE expander. github.com/teambition/rrule-go
// is the only credible one; its last release is v1.8.2 of 2023-01-13, it reads
// the system clock internally when DTSTART is unset — which would put a hidden,
// non-deterministic clock read inside a package whose no-clock rule is enforced
// mechanically — and it expands to zoned instants, while habit_checks is keyed
// by date and every result would have to be folded back to one. The maintained
// iCalendar libraries (github.com/arran4/golang-ical, github.com/vareversat/gics)
// parse calendars but do not expand recurrences at all.
//
// What habits actually need is small and closed, so it is implemented here,
// exactly, and everything else is refused:
//
//	FREQ=DAILY
//	FREQ=WEEKLY    with optional BYDAY (MO,TU,WE,TH,FR,SA,SU)
//	FREQ=MONTHLY   with optional BYMONTHDAY (1..31)
//	INTERVAL=n     (n >= 1), on any of the three
//	WKST=MO        accepted, since it is the default this implementation assumes
//
// Anything else — COUNT, UNTIL, BYSETPOS, BYMONTH, BYWEEKNO, BYYEARDAY, an
// ordinal BYDAY such as 2MO, a negative BYMONTHDAY such as -1, or any other
// FREQ — is an ErrUnsupportedRecurrence from ParseRecurrence. Refusing at parse
// time is the whole point: a silently mis-expanded rule produces a wrong streak
// that looks right.
//
// # Timezones
//
// There are none. Habits are date-granular because habit_checks is keyed by
// date, so expansion works in Dates throughout and there is no instant, no zone
// and no hour anywhere in this file.
type Recurrence struct {
	Freq     Frequency
	Interval int
	// ByDay applies to FREQ=WEEKLY. Empty means "the weekday DTSTART falls on",
	// which is what RFC 5545 says an absent BYDAY means.
	ByDay []time.Weekday
	// ByMonthDay applies to FREQ=MONTHLY, values 1..31. Empty means "the day of
	// the month DTSTART falls on". A day a given month does not have is simply
	// skipped, so BYMONTHDAY=31 fires seven times a year and BYMONTHDAY=29
	// fires in February only in leap years.
	ByMonthDay []int
}

// ParseRecurrence parses an RRULE restricted to the supported subset. An
// optional leading "RRULE:" is accepted; keys and values are case-insensitive.
func ParseRecurrence(rule string) (Recurrence, error) {
	raw := strings.TrimSpace(rule)
	raw = strings.TrimPrefix(strings.ToUpper(raw), "RRULE:")
	if raw == "" {
		return Recurrence{}, fmt.Errorf("domain: empty recurrence rule: %w", ErrUnsupportedRecurrence)
	}

	r := Recurrence{Interval: 1}
	seen := make(map[string]bool)
	freqSet := false

	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			return Recurrence{}, fmt.Errorf("domain: empty part in %q: %w", rule, ErrUnsupportedRecurrence)
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || value == "" {
			return Recurrence{}, fmt.Errorf("domain: %q is not KEY=VALUE: %w", part, ErrUnsupportedRecurrence)
		}
		if seen[key] {
			return Recurrence{}, fmt.Errorf("domain: %s given twice: %w", key, ErrUnsupportedRecurrence)
		}
		seen[key] = true

		switch key {
		case "FREQ":
			switch Frequency(value) {
			case FreqDaily, FreqWeekly, FreqMonthly:
				r.Freq = Frequency(value)
				freqSet = true
			default:
				return Recurrence{}, fmt.Errorf(
					"domain: FREQ=%s is not one of DAILY, WEEKLY, MONTHLY: %w", value, ErrUnsupportedRecurrence)
			}

		case "INTERVAL":
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return Recurrence{}, fmt.Errorf(
					"domain: INTERVAL=%s is not a positive integer: %w", value, ErrUnsupportedRecurrence)
			}
			r.Interval = n

		case "BYDAY":
			days, err := parseByDay(value)
			if err != nil {
				return Recurrence{}, err
			}
			r.ByDay = days

		case "BYMONTHDAY":
			days, err := parseByMonthDay(value)
			if err != nil {
				return Recurrence{}, err
			}
			r.ByMonthDay = days

		case "WKST":
			// The week arithmetic here is Monday-based, which is RFC 5545's
			// default. Any other week start would change which occurrences
			// FREQ=WEEKLY;INTERVAL>1 lands on, so it is refused rather than
			// ignored.
			if value != "MO" {
				return Recurrence{}, fmt.Errorf(
					"domain: WKST=%s is not supported, only WKST=MO: %w", value, ErrUnsupportedRecurrence)
			}

		default:
			return Recurrence{}, fmt.Errorf(
				"domain: %s is not a supported RRULE part: %w", key, ErrUnsupportedRecurrence)
		}
	}

	if !freqSet {
		return Recurrence{}, fmt.Errorf("domain: %q has no FREQ: %w", rule, ErrUnsupportedRecurrence)
	}
	if len(r.ByDay) > 0 && r.Freq != FreqWeekly {
		return Recurrence{}, fmt.Errorf(
			"domain: BYDAY is only supported with FREQ=WEEKLY, not %s: %w", r.Freq, ErrUnsupportedRecurrence)
	}
	if len(r.ByMonthDay) > 0 && r.Freq != FreqMonthly {
		return Recurrence{}, fmt.Errorf(
			"domain: BYMONTHDAY is only supported with FREQ=MONTHLY, not %s: %w", r.Freq, ErrUnsupportedRecurrence)
	}
	return r, nil
}

var weekdayCodes = map[string]time.Weekday{
	"SU": time.Sunday,
	"MO": time.Monday,
	"TU": time.Tuesday,
	"WE": time.Wednesday,
	"TH": time.Thursday,
	"FR": time.Friday,
	"SA": time.Saturday,
}

func parseByDay(value string) ([]time.Weekday, error) {
	var out []time.Weekday
	for _, code := range strings.Split(value, ",") {
		code = strings.TrimSpace(code)
		wd, ok := weekdayCodes[code]
		if !ok {
			// This is where "2MO" and "-1FR" land: an ordinal BYDAY means "the
			// second Monday of the period", which this implementation does not
			// do, so it must not be mistaken for a plain Monday.
			return nil, fmt.Errorf(
				"domain: BYDAY=%s is not a plain two-letter weekday: %w", code, ErrUnsupportedRecurrence)
		}
		for _, already := range out {
			if already == wd {
				return nil, fmt.Errorf("domain: BYDAY lists %s twice: %w", code, ErrUnsupportedRecurrence)
			}
		}
		out = append(out, wd)
	}
	return out, nil
}

func parseByMonthDay(value string) ([]int, error) {
	var out []int
	for _, field := range strings.Split(value, ",") {
		field = strings.TrimSpace(field)
		n, err := strconv.Atoi(field)
		if err != nil || n < 1 || n > 31 {
			// Negative BYMONTHDAY ("-1" = the last day of the month) is a real
			// RFC 5545 construct and is NOT implemented, so it is refused here
			// rather than silently treated as a positive day.
			return nil, fmt.Errorf(
				"domain: BYMONTHDAY=%s is not a day in 1..31: %w", field, ErrUnsupportedRecurrence)
		}
		for _, already := range out {
			if already == n {
				return nil, fmt.Errorf("domain: BYMONTHDAY lists %d twice: %w", n, ErrUnsupportedRecurrence)
			}
		}
		out = append(out, n)
	}
	return out, nil
}

// Matches reports whether d is a scheduled occurrence of r anchored at dtstart.
//
// Membership is computed from d and dtstart alone rather than by counting
// forward from dtstart, so asking about a date a decade out costs the same as
// asking about tomorrow, and no iteration can drift.
func (r Recurrence) Matches(dtstart, d Date) bool {
	if d.Before(dtstart) {
		return false
	}

	switch r.Freq {
	case FreqDaily:
		return dtstart.DaysUntil(d)%r.Interval == 0

	case FreqWeekly:
		if weeksBetween(dtstart, d)%r.Interval != 0 {
			return false
		}
		if len(r.ByDay) == 0 {
			return d.Weekday() == dtstart.Weekday()
		}
		for _, wd := range r.ByDay {
			if d.Weekday() == wd {
				return true
			}
		}
		return false

	case FreqMonthly:
		if monthsBetween(dtstart, d)%r.Interval != 0 {
			return false
		}
		if len(r.ByMonthDay) == 0 {
			return d.Day == dtstart.Day
		}
		for _, md := range r.ByMonthDay {
			if d.Day == md {
				return true
			}
		}
		return false

	default:
		// Unreachable through ParseRecurrence, which is the only way to build a
		// Recurrence with a FREQ. A zero-value Recurrence matches nothing,
		// which is the safe answer.
		return false
	}
}

// mondayOf returns the Monday of d's week. RFC 5545's default week start is
// Monday and WKST is restricted to MO, so this is the only week boundary in
// play.
func mondayOf(d Date) Date {
	return d.AddDays(-((int(d.Weekday()) + 6) % 7))
}

// weeksBetween counts whole weeks between the week containing a and the week
// containing b. Counting between week boundaries rather than between the dates
// is what makes FREQ=WEEKLY;INTERVAL=2 mean "every other week" rather than
// "every fourteenth day from DTSTART", which is what it would mean if the days
// were divided directly and which differs as soon as BYDAY moves an occurrence
// off DTSTART's weekday.
func weeksBetween(a, b Date) int {
	return mondayOf(a).DaysUntil(mondayOf(b)) / 7
}

// monthsBetween counts whole calendar months from a to b.
func monthsBetween(a, b Date) int {
	return (b.Year-a.Year)*12 + int(b.Month) - int(a.Month)
}

// Expand returns every occurrence of r in [from, to] inclusive, ascending.
// Occurrences before dtstart are never produced.
func (r Recurrence) Expand(dtstart, from, to Date) ([]Date, error) {
	start := from
	if start.Before(dtstart) {
		start = dtstart
	}
	if start.After(to) {
		return nil, nil
	}
	if span := start.DaysUntil(to); span > maxWindowDays {
		return nil, fmt.Errorf("domain: %d days from %s to %s (max %d): %w",
			span, start, to, maxWindowDays, ErrWindowTooLarge)
	}

	var out []Date
	for d := start; !d.After(to); d = d.AddDays(1) {
		if r.Matches(dtstart, d) {
			out = append(out, d)
		}
	}
	return out, nil
}

// Next returns the first occurrence strictly after `after`, searching at most
// maxSearchDays forward. ok is false when the rule produces nothing in that
// horizon.
func (r Recurrence) Next(dtstart, after Date) (Date, bool) {
	d := after.AddDays(1)
	if d.Before(dtstart) {
		d = dtstart
	}
	for range maxSearchDays {
		if r.Matches(dtstart, d) {
			return d, true
		}
		d = d.AddDays(1)
	}
	return Date{}, false
}

// Previous returns the last occurrence strictly before `before`, searching back
// at most maxSearchDays and never earlier than dtstart. ok is false when there
// is none.
func (r Recurrence) Previous(dtstart, before Date) (Date, bool) {
	d := before.AddDays(-1)
	for range maxSearchDays {
		if d.Before(dtstart) {
			return Date{}, false
		}
		if r.Matches(dtstart, d) {
			return d, true
		}
		d = d.AddDays(-1)
	}
	return Date{}, false
}

// Occurrences parses rule and returns every occurrence in [from, to] inclusive,
// ascending. This is the entry point S1-12 consumes.
//
// The window is inclusive at both ends and bounded, so an open-ended rule always
// returns a bounded result. An unparseable or unsupported rule is an error, never
// an empty slice: "no occurrences" and "I did not understand the rule" are
// different answers and a streak of 0 must not be able to mean the second one.
func Occurrences(rule string, dtstart, from, to Date) ([]Date, error) {
	r, err := ParseRecurrence(rule)
	if err != nil {
		return nil, err
	}
	return r.Expand(dtstart, from, to)
}

// NextOccurrence returns the first occurrence of rule strictly after `after`.
func NextOccurrence(rule string, dtstart, after Date) (Date, bool, error) {
	r, err := ParseRecurrence(rule)
	if err != nil {
		return Date{}, false, err
	}
	d, ok := r.Next(dtstart, after)
	return d, ok, nil
}

// PreviousOccurrence returns the last occurrence of rule strictly before
// `before`, never earlier than dtstart.
func PreviousOccurrence(rule string, dtstart, before Date) (Date, bool, error) {
	r, err := ParseRecurrence(rule)
	if err != nil {
		return Date{}, false, err
	}
	d, ok := r.Previous(dtstart, before)
	return d, ok, nil
}
