package domain_test

import (
	"testing"
	"time"

	"nexus/internal/domain"
)

// clockAt returns an injected clock reporting 10:30 on the given date, in the
// zone given. The domain never reads the wall clock; every date below is
// computed from one of these.
func clockAt(d domain.Date, loc *time.Location) func() time.Time {
	return func() time.Time {
		return time.Date(d.Year, d.Month, d.Day, 10, 30, 0, 0, loc)
	}
}

func utcClockAt(d domain.Date) func() time.Time { return clockAt(d, time.UTC) }

func dueNode(status domain.Status, due *domain.Date, src domain.DueSource) domain.Node {
	n := task("n", "", status)
	n.Due = due
	n.DueSource = src
	return n
}

func TestToday(t *testing.T) {
	want := domain.NewDate(2026, time.September, 18)
	if got := domain.Today(utcClockAt(want)); got != want {
		t.Errorf("Today() = %v, want %v", got, want)
	}
}

// Today is the user's today: the calendar day in the clock's own zone, never
// folded through UTC first.
func TestTodayUsesTheClocksOwnZone(t *testing.T) {
	east := time.FixedZone("UTC+13", 13*60*60)
	now := func() time.Time {
		// 01:00 on the 19th in UTC+13 is still 12:00 on the 18th in UTC.
		return time.Date(2026, time.September, 19, 1, 0, 0, 0, east)
	}
	if got, want := domain.Today(now), domain.NewDate(2026, time.September, 19); got != want {
		t.Errorf("Today() = %v, want %v (the user's day, not UTC's)", got, want)
	}
}

// D1: → This week is the upcoming Friday, and Friday maps to itself. Every
// weekday gets a row.
func TestUpcomingFriday(t *testing.T) {
	tests := []struct {
		name        string
		today       domain.Date
		wantWeekday time.Weekday
		want        domain.Date
	}{
		{"monday", domain.NewDate(2026, time.September, 14), time.Monday, domain.NewDate(2026, time.September, 18)},
		{"tuesday", domain.NewDate(2026, time.September, 15), time.Tuesday, domain.NewDate(2026, time.September, 18)},
		{"wednesday", domain.NewDate(2026, time.September, 16), time.Wednesday, domain.NewDate(2026, time.September, 18)},
		{"thursday", domain.NewDate(2026, time.September, 17), time.Thursday, domain.NewDate(2026, time.September, 18)},
		{"friday is today itself", domain.NewDate(2026, time.September, 18), time.Friday, domain.NewDate(2026, time.September, 18)},
		{"saturday is six days out", domain.NewDate(2026, time.September, 19), time.Saturday, domain.NewDate(2026, time.September, 25)},
		{"sunday", domain.NewDate(2026, time.September, 20), time.Sunday, domain.NewDate(2026, time.September, 25)},

		// Month end: 2026-10-31 is a Saturday, so the answer crosses into
		// November.
		{"month end saturday", domain.NewDate(2026, time.October, 31), time.Saturday, domain.NewDate(2026, time.November, 6)},
		// Year end: 2026-12-31 is a Thursday, so the answer crosses into 2027.
		{"year end thursday", domain.NewDate(2026, time.December, 31), time.Thursday, domain.NewDate(2027, time.January, 1)},
		{"christmas day is a friday", domain.NewDate(2026, time.December, 25), time.Friday, domain.NewDate(2026, time.December, 25)},
		// A leap year, for the February arithmetic.
		{"leap day is a tuesday", domain.NewDate(2028, time.February, 29), time.Tuesday, domain.NewDate(2028, time.March, 3)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Self-check the fixture: if the calendar assumption in the row is
			// wrong, the expectation is meaningless.
			if got := tt.today.Weekday(); got != tt.wantWeekday {
				t.Fatalf("fixture is wrong: %v is a %v, not a %v", tt.today, got, tt.wantWeekday)
			}
			got := domain.UpcomingFriday(tt.today)
			if got != tt.want {
				t.Fatalf("UpcomingFriday(%v) = %v, want %v", tt.today, got, tt.want)
			}
			if got.Weekday() != time.Friday {
				t.Errorf("UpcomingFriday(%v) = %v, which is a %v", tt.today, got, got.Weekday())
			}
			if got.Before(tt.today) {
				t.Errorf("UpcomingFriday(%v) = %v, which is in the past", tt.today, got)
			}
			if d := tt.today.DaysUntil(got); d < 0 || d > 6 {
				t.Errorf("UpcomingFriday(%v) is %d days out, want 0..6", tt.today, d)
			}
		})
	}
}

// D1, every column, over every weekday for the week target.
func TestDueForColumnMove(t *testing.T) {
	manualDate := domain.NewDate(2026, time.March, 1)

	tests := []struct {
		name    string
		today   domain.Date
		node    domain.Node
		target  domain.Status
		wantDue *domain.Date
		wantSrc domain.DueSource
	}{
		{
			name:    "to today sets the date and the auto source",
			today:   domain.NewDate(2026, time.September, 18),
			node:    dueNode(domain.StatusBacklog, nil, domain.DueSourceManual),
			target:  domain.StatusToday,
			wantDue: ptr(domain.NewDate(2026, time.September, 18)),
			wantSrc: domain.DueSourceAuto,
		},
		{
			name:    "to week from a monday sets the friday",
			today:   domain.NewDate(2026, time.September, 14),
			node:    dueNode(domain.StatusBacklog, nil, domain.DueSourceManual),
			target:  domain.StatusWeek,
			wantDue: ptr(domain.NewDate(2026, time.September, 18)),
			wantSrc: domain.DueSourceAuto,
		},
		{
			name:    "to week on a friday sets today itself",
			today:   domain.NewDate(2026, time.September, 18),
			node:    dueNode(domain.StatusBacklog, nil, domain.DueSourceManual),
			target:  domain.StatusWeek,
			wantDue: ptr(domain.NewDate(2026, time.September, 18)),
			wantSrc: domain.DueSourceAuto,
		},
		{
			name:    "to week on a saturday sets the friday six days out",
			today:   domain.NewDate(2026, time.September, 19),
			node:    dueNode(domain.StatusBacklog, nil, domain.DueSourceManual),
			target:  domain.StatusWeek,
			wantDue: ptr(domain.NewDate(2026, time.September, 25)),
			wantSrc: domain.DueSourceAuto,
		},
		{
			name:    "to backlog clears an auto date",
			today:   domain.NewDate(2026, time.September, 18),
			node:    dueNode(domain.StatusToday, ptr(domain.NewDate(2026, time.September, 18)), domain.DueSourceAuto),
			target:  domain.StatusBacklog,
			wantDue: nil,
			wantSrc: domain.DueSourceManual,
		},
		{
			name:    "to backlog keeps a manual date and its manual source",
			today:   domain.NewDate(2026, time.September, 18),
			node:    dueNode(domain.StatusToday, &manualDate, domain.DueSourceManual),
			target:  domain.StatusBacklog,
			wantDue: &manualDate,
			wantSrc: domain.DueSourceManual,
		},
		{
			name:    "to backlog with no date at all is a no-op",
			today:   domain.NewDate(2026, time.September, 18),
			node:    dueNode(domain.StatusToday, nil, domain.DueSourceManual),
			target:  domain.StatusBacklog,
			wantDue: nil,
			wantSrc: domain.DueSourceManual,
		},
		{
			name:    "to doing leaves an auto date untouched",
			today:   domain.NewDate(2026, time.September, 18),
			node:    dueNode(domain.StatusToday, ptr(domain.NewDate(2026, time.September, 18)), domain.DueSourceAuto),
			target:  domain.StatusDoing,
			wantDue: ptr(domain.NewDate(2026, time.September, 18)),
			wantSrc: domain.DueSourceAuto,
		},
		{
			name:    "to doing leaves a manual date untouched",
			today:   domain.NewDate(2026, time.September, 18),
			node:    dueNode(domain.StatusToday, &manualDate, domain.DueSourceManual),
			target:  domain.StatusDoing,
			wantDue: &manualDate,
			wantSrc: domain.DueSourceManual,
		},
		{
			name:    "to done leaves a manual date untouched",
			today:   domain.NewDate(2026, time.September, 18),
			node:    dueNode(domain.StatusDoing, &manualDate, domain.DueSourceManual),
			target:  domain.StatusDone,
			wantDue: &manualDate,
			wantSrc: domain.DueSourceManual,
		},
		{
			name:    "to done leaves a nil date nil",
			today:   domain.NewDate(2026, time.September, 18),
			node:    dueNode(domain.StatusDoing, nil, domain.DueSourceAuto),
			target:  domain.StatusDone,
			wantDue: nil,
			wantSrc: domain.DueSourceAuto,
		},
		{
			name:    "an unknown target changes nothing",
			today:   domain.NewDate(2026, time.September, 18),
			node:    dueNode(domain.StatusDoing, &manualDate, domain.DueSourceManual),
			target:  domain.Status("archived"),
			wantDue: &manualDate,
			wantSrc: domain.DueSourceManual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.DueForColumnMove(tt.node, tt.target, utcClockAt(tt.today))
			assertDueUpdate(t, got, tt.wantDue, tt.wantSrc)
		})
	}
}

// D8, the reading of D1 confirmed by the user: a column move to Today or
// This-week ALWAYS overwrites the existing due date, including a manual one, and
// flips due_source to auto. The column move always wins, so the column and the
// date can never disagree.
func TestColumnMoveOverwritesAManualDueDate(t *testing.T) {
	hand := domain.NewDate(2026, time.March, 1)
	today := domain.NewDate(2026, time.September, 18)
	now := utcClockAt(today)

	t.Run("to today", func(t *testing.T) {
		n := dueNode(domain.StatusBacklog, &hand, domain.DueSourceManual)
		got := domain.DueForColumnMove(n, domain.StatusToday, now)
		assertDueUpdate(t, got, &today, domain.DueSourceAuto)
	})

	t.Run("to week", func(t *testing.T) {
		n := dueNode(domain.StatusBacklog, &hand, domain.DueSourceManual)
		got := domain.DueForColumnMove(n, domain.StatusWeek, now)
		// 2026-09-18 is a Friday, so the upcoming Friday is today.
		assertDueUpdate(t, got, &today, domain.DueSourceAuto)
	})

	// The accepted consequence, stated as a test: once the drag has made the
	// date auto, a later move to Backlog deletes it entirely.
	t.Run("and a later move to backlog then clears it entirely", func(t *testing.T) {
		n := dueNode(domain.StatusBacklog, &hand, domain.DueSourceManual)
		n = domain.DueForColumnMove(n, domain.StatusToday, now).ApplyTo(n)
		n.Status = domain.StatusToday

		got := domain.DueForColumnMove(n, domain.StatusBacklog, now)
		assertDueUpdate(t, got, nil, domain.DueSourceManual)
	})
}

// D1: any user edit of due is manual, including setting it to the value it
// already had and including clearing it.
func TestDueForUserEdit(t *testing.T) {
	d := domain.NewDate(2026, time.September, 30)

	tests := []struct {
		name    string
		in      *domain.Date
		wantDue *domain.Date
	}{
		{"a date the user typed", &d, &d},
		{"clearing the date", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.DueForUserEdit(tt.in)
			assertDueUpdate(t, got, tt.wantDue, domain.DueSourceManual)
		})
	}

	t.Run("re-typing the same auto date makes it manual", func(t *testing.T) {
		auto := domain.NewDate(2026, time.September, 18)
		n := dueNode(domain.StatusToday, &auto, domain.DueSourceAuto)

		got := domain.DueForUserEdit(n.Due)
		assertDueUpdate(t, got, &auto, domain.DueSourceManual)

		// And it now survives a move to Backlog, which is the whole point.
		n = got.ApplyTo(n)
		back := domain.DueForColumnMove(n, domain.StatusBacklog, utcClockAt(auto))
		assertDueUpdate(t, back, &auto, domain.DueSourceManual)
	})
}

// The update must not alias the caller's Date, or two nodes could end up sharing
// one and editing each other's due date.
func TestDueUpdateDoesNotAliasTheCallersDate(t *testing.T) {
	d := domain.NewDate(2026, time.September, 18)
	got := domain.DueForUserEdit(&d)
	if got.Due == &d {
		t.Fatal("DueForUserEdit returned the caller's pointer")
	}

	n := domain.DueUpdate{Due: &d, DueSource: domain.DueSourceManual}.ApplyTo(task("n", "", domain.StatusBacklog))
	if n.Due == &d {
		t.Fatal("ApplyTo stored the caller's pointer")
	}
	*n.Due = domain.NewDate(1999, time.January, 1)
	if d != domain.NewDate(2026, time.September, 18) {
		t.Errorf("mutating the node's due date changed the caller's: %v", d)
	}
}

func TestDueUpdateApplyTo(t *testing.T) {
	d := domain.NewDate(2026, time.September, 18)
	before := task("n", "", domain.StatusBacklog)
	before.Due = ptr(domain.NewDate(2020, time.January, 1))
	before.DueSource = domain.DueSourceManual

	after := domain.DueUpdate{Due: &d, DueSource: domain.DueSourceAuto}.ApplyTo(before)

	if after.Due == nil || *after.Due != d {
		t.Errorf("Due = %v, want %v", after.Due, d)
	}
	if after.DueSource != domain.DueSourceAuto {
		t.Errorf("DueSource = %q, want auto", after.DueSource)
	}
	// ApplyTo works on a copy: the original is untouched.
	if before.Due == nil || *before.Due != domain.NewDate(2020, time.January, 1) {
		t.Errorf("ApplyTo mutated its argument: %v", before.Due)
	}
	if before.DueSource != domain.DueSourceManual {
		t.Errorf("ApplyTo mutated its argument's due_source: %q", before.DueSource)
	}

	cleared := domain.DueUpdate{Due: nil, DueSource: domain.DueSourceManual}.ApplyTo(after)
	if cleared.Due != nil {
		t.Errorf("Due = %v, want nil", cleared.Due)
	}
}

// D1: overdue is due < today AND status != done. Equal to today is not overdue.
func TestIsOverdue(t *testing.T) {
	today := domain.NewDate(2026, time.September, 18)
	yesterday := today.AddDays(-1)
	tomorrow := today.AddDays(1)

	tests := []struct {
		name   string
		due    *domain.Date
		status domain.Status
		want   bool
	}{
		{"yesterday and today status", &yesterday, domain.StatusToday, true},
		{"yesterday and backlog", &yesterday, domain.StatusBacklog, true},
		{"yesterday and doing", &yesterday, domain.StatusDoing, true},
		{"last year and week", ptr(domain.NewDate(2025, time.December, 31)), domain.StatusWeek, true},
		{"due exactly today is not overdue", &today, domain.StatusToday, false},
		{"tomorrow is not overdue", &tomorrow, domain.StatusToday, false},
		{"yesterday but done is not overdue", &yesterday, domain.StatusDone, false},
		{"last year but done is not overdue", ptr(domain.NewDate(2025, time.December, 31)), domain.StatusDone, false},
		{"no due date is never overdue", nil, domain.StatusBacklog, false},
		{"no due date and done is never overdue", nil, domain.StatusDone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := dueNode(tt.status, tt.due, domain.DueSourceManual)
			if got := domain.IsOverdue(n, today); got != tt.want {
				t.Errorf("IsOverdue(due=%v, status=%q) = %v, want %v", tt.due, tt.status, got, tt.want)
			}
		})
	}
}

func assertDueUpdate(t *testing.T, got domain.DueUpdate, wantDue *domain.Date, wantSrc domain.DueSource) {
	t.Helper()

	switch {
	case wantDue == nil && got.Due != nil:
		t.Errorf("Due = %v, want nil", got.Due)
	case wantDue != nil && got.Due == nil:
		t.Errorf("Due = nil, want %v", *wantDue)
	case wantDue != nil && *got.Due != *wantDue:
		t.Errorf("Due = %v, want %v", *got.Due, *wantDue)
	}
	if got.DueSource != wantSrc {
		t.Errorf("DueSource = %q, want %q", got.DueSource, wantSrc)
	}
}
