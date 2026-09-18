package domain_test

import (
	"errors"
	"testing"
	"time"

	"nexus/internal/domain"
)

func d(y int, m time.Month, day int) domain.Date { return domain.NewDate(y, m, day) }

func dateStrings(ds []domain.Date) []string {
	out := make([]string, len(ds))
	for i, x := range ds {
		out[i] = x.String()
	}
	return out
}

func TestOccurrences(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		dtstart  domain.Date
		from, to domain.Date
		want     []string
	}{
		{
			name: "FREQ=DAILY",
			rule: "FREQ=DAILY", dtstart: d(2026, time.September, 14),
			from: d(2026, time.September, 14), to: d(2026, time.September, 18),
			want: []string{"2026-09-14", "2026-09-15", "2026-09-16", "2026-09-17", "2026-09-18"},
		},
		{
			name: "FREQ=DAILY;INTERVAL=3",
			rule: "FREQ=DAILY;INTERVAL=3", dtstart: d(2026, time.September, 14),
			from: d(2026, time.September, 14), to: d(2026, time.September, 24),
			want: []string{"2026-09-14", "2026-09-17", "2026-09-20", "2026-09-23"},
		},
		{
			name: "FREQ=WEEKLY plain repeats DTSTART's weekday",
			rule: "FREQ=WEEKLY", dtstart: d(2026, time.September, 18), // a Friday
			from: d(2026, time.September, 1), to: d(2026, time.October, 16),
			want: []string{"2026-09-18", "2026-09-25", "2026-10-02", "2026-10-09", "2026-10-16"},
		},
		{
			name: "FREQ=WEEKLY;BYDAY=MO,WE,FR",
			rule: "FREQ=WEEKLY;BYDAY=MO,WE,FR", dtstart: d(2026, time.September, 14), // a Monday
			from: d(2026, time.September, 14), to: d(2026, time.September, 25),
			want: []string{
				"2026-09-14", "2026-09-16", "2026-09-18",
				"2026-09-21", "2026-09-23", "2026-09-25",
			},
		},
		{
			name: "FREQ=WEEKLY;INTERVAL=2 skips the odd weeks",
			rule: "FREQ=WEEKLY;INTERVAL=2", dtstart: d(2026, time.September, 18),
			from: d(2026, time.September, 1), to: d(2026, time.November, 1),
			want: []string{"2026-09-18", "2026-10-02", "2026-10-16", "2026-10-30"},
		},
		{
			name: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,FR keeps both days in the on-weeks",
			rule: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,FR", dtstart: d(2026, time.September, 14),
			from: d(2026, time.September, 14), to: d(2026, time.October, 12),
			want: []string{"2026-09-14", "2026-09-18", "2026-09-28", "2026-10-02", "2026-10-12"},
		},
		{
			// The interval counts WEEKS, not days divided by seven. DTSTART is
			// a Friday, so Monday 09-21 is in the NEXT week — week 1, an off
			// week — even though it is only three days after DTSTART. Counting
			// (d - dtstart)/7 would put it in week 0 and schedule it.
			name: "FREQ=WEEKLY;INTERVAL=2 counts from the week boundary, not from DTSTART",
			rule: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,FR", dtstart: d(2026, time.September, 18), // a Friday
			from: d(2026, time.September, 14), to: d(2026, time.October, 5),
			want: []string{"2026-09-18", "2026-09-28", "2026-10-02"},
		},
		{
			name: "FREQ=MONTHLY;BYMONTHDAY=1",
			rule: "FREQ=MONTHLY;BYMONTHDAY=1", dtstart: d(2026, time.September, 1),
			from: d(2026, time.September, 1), to: d(2027, time.January, 31),
			want: []string{"2026-09-01", "2026-10-01", "2026-11-01", "2026-12-01", "2027-01-01"},
		},
		{
			name: "FREQ=MONTHLY plain repeats DTSTART's day of the month",
			rule: "FREQ=MONTHLY", dtstart: d(2026, time.September, 18),
			from: d(2026, time.September, 1), to: d(2026, time.December, 31),
			want: []string{"2026-09-18", "2026-10-18", "2026-11-18", "2026-12-18"},
		},
		{
			name: "FREQ=MONTHLY;INTERVAL=3",
			rule: "FREQ=MONTHLY;INTERVAL=3", dtstart: d(2026, time.January, 15),
			from: d(2026, time.January, 1), to: d(2026, time.December, 31),
			want: []string{"2026-01-15", "2026-04-15", "2026-07-15", "2026-10-15"},
		},
		{
			// Month end: the 31st simply does not exist in 30-day months, so
			// those months are skipped rather than rolled into the 1st.
			name: "FREQ=MONTHLY;BYMONTHDAY=31 skips the short months",
			rule: "FREQ=MONTHLY;BYMONTHDAY=31", dtstart: d(2026, time.January, 31),
			from: d(2026, time.January, 1), to: d(2026, time.August, 31),
			want: []string{"2026-01-31", "2026-03-31", "2026-05-31", "2026-07-31", "2026-08-31"},
		},
		{
			// Leap year: 29 February exists in 2028 and not in 2026 or 2027.
			name: "FREQ=MONTHLY;BYMONTHDAY=29 lands on a leap day only in a leap year",
			rule: "FREQ=MONTHLY;BYMONTHDAY=29", dtstart: d(2027, time.December, 29),
			from: d(2028, time.January, 1), to: d(2028, time.April, 30),
			want: []string{"2028-01-29", "2028-02-29", "2028-03-29", "2028-04-29"},
		},
		{
			name: "FREQ=MONTHLY;BYMONTHDAY=29 finds no February in a common year",
			rule: "FREQ=MONTHLY;BYMONTHDAY=29", dtstart: d(2025, time.December, 29),
			from: d(2026, time.February, 1), to: d(2026, time.February, 28),
			want: nil,
		},
		{
			name: "several BYMONTHDAY values",
			rule: "FREQ=MONTHLY;BYMONTHDAY=1,15", dtstart: d(2026, time.September, 1),
			from: d(2026, time.September, 1), to: d(2026, time.October, 31),
			want: []string{"2026-09-01", "2026-09-15", "2026-10-01", "2026-10-15"},
		},
		{
			name: "the window clips the front and the back",
			rule: "FREQ=DAILY", dtstart: d(2026, time.September, 1),
			from: d(2026, time.September, 10), to: d(2026, time.September, 12),
			want: []string{"2026-09-10", "2026-09-11", "2026-09-12"},
		},
		{
			name: "nothing before DTSTART, whatever the window asks for",
			rule: "FREQ=DAILY", dtstart: d(2026, time.September, 10),
			from: d(2026, time.January, 1), to: d(2026, time.September, 12),
			want: []string{"2026-09-10", "2026-09-11", "2026-09-12"},
		},
		{
			name: "a window entirely before DTSTART is empty",
			rule: "FREQ=DAILY", dtstart: d(2026, time.September, 10),
			from: d(2026, time.January, 1), to: d(2026, time.January, 31),
			want: nil,
		},
		{
			name: "a window that ends before it starts is empty",
			rule: "FREQ=DAILY", dtstart: d(2026, time.September, 1),
			from: d(2026, time.September, 10), to: d(2026, time.September, 5),
			want: nil,
		},
		{
			name: "a single-day window on an occurrence",
			rule: "FREQ=WEEKLY;BYDAY=FR", dtstart: d(2026, time.September, 14),
			from: d(2026, time.September, 18), to: d(2026, time.September, 18),
			want: []string{"2026-09-18"},
		},
		{
			name: "a BYDAY before DTSTART's weekday is skipped in the first week",
			rule: "FREQ=WEEKLY;BYDAY=MO,FR", dtstart: d(2026, time.September, 18), // a Friday
			from: d(2026, time.September, 1), to: d(2026, time.September, 30),
			want: []string{"2026-09-18", "2026-09-21", "2026-09-25", "2026-09-28"},
		},
		{
			name: "an RRULE: prefix and lower case are accepted",
			rule: "rrule:freq=weekly;byday=mo,we,fr", dtstart: d(2026, time.September, 14),
			from: d(2026, time.September, 14), to: d(2026, time.September, 18),
			want: []string{"2026-09-14", "2026-09-16", "2026-09-18"},
		},
		{
			name: "WKST=MO is accepted",
			rule: "FREQ=WEEKLY;INTERVAL=2;WKST=MO", dtstart: d(2026, time.September, 18),
			from: d(2026, time.September, 18), to: d(2026, time.October, 3),
			want: []string{"2026-09-18", "2026-10-02"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.Occurrences(tt.rule, tt.dtstart, tt.from, tt.to)
			if err != nil {
				t.Fatalf("Occurrences() = %v", err)
			}
			if !equalStrings(dateStrings(got), tt.want) {
				t.Errorf("Occurrences(%q) = %v, want %v", tt.rule, dateStrings(got), tt.want)
			}
		})
	}
}

// Anything outside the supported subset is refused at parse time with a
// matchable error — never an empty slice and never a partial expansion.
func TestParseRecurrenceRejectsWhatItCannotExpand(t *testing.T) {
	tests := []struct {
		name string
		rule string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"just the prefix", "RRULE:"},
		{"not an rrule at all", "every monday"},
		{"no FREQ", "INTERVAL=2"},
		{"FREQ=YEARLY", "FREQ=YEARLY"},
		{"FREQ=HOURLY", "FREQ=HOURLY"},
		{"FREQ=MINUTELY", "FREQ=MINUTELY"},
		{"FREQ=SECONDLY", "FREQ=SECONDLY"},
		{"an unknown FREQ", "FREQ=FORTNIGHTLY"},
		{"COUNT", "FREQ=DAILY;COUNT=10"},
		{"UNTIL", "FREQ=DAILY;UNTIL=20261231T000000Z"},
		{"BYSETPOS", "FREQ=MONTHLY;BYDAY=MO;BYSETPOS=-1"},
		{"BYMONTH", "FREQ=MONTHLY;BYMONTH=2"},
		{"BYWEEKNO", "FREQ=WEEKLY;BYWEEKNO=12"},
		{"BYYEARDAY", "FREQ=DAILY;BYYEARDAY=100"},
		{"BYHOUR", "FREQ=DAILY;BYHOUR=9"},
		{"an ordinal BYDAY", "FREQ=WEEKLY;BYDAY=2MO"},
		{"a negative ordinal BYDAY", "FREQ=WEEKLY;BYDAY=-1FR"},
		{"an unknown weekday code", "FREQ=WEEKLY;BYDAY=XX"},
		{"a three-letter weekday", "FREQ=WEEKLY;BYDAY=MON"},
		{"a negative BYMONTHDAY", "FREQ=MONTHLY;BYMONTHDAY=-1"},
		{"a zero BYMONTHDAY", "FREQ=MONTHLY;BYMONTHDAY=0"},
		{"a BYMONTHDAY past 31", "FREQ=MONTHLY;BYMONTHDAY=32"},
		{"a non-numeric BYMONTHDAY", "FREQ=MONTHLY;BYMONTHDAY=last"},
		{"INTERVAL=0", "FREQ=DAILY;INTERVAL=0"},
		{"a negative INTERVAL", "FREQ=DAILY;INTERVAL=-2"},
		{"a non-numeric INTERVAL", "FREQ=DAILY;INTERVAL=two"},
		{"WKST=SU", "FREQ=WEEKLY;WKST=SU"},
		{"BYDAY on a daily rule", "FREQ=DAILY;BYDAY=MO"},
		{"BYDAY on a monthly rule", "FREQ=MONTHLY;BYDAY=MO"},
		{"BYMONTHDAY on a weekly rule", "FREQ=WEEKLY;BYMONTHDAY=1"},
		{"a part with no equals sign", "FREQ=DAILY;NONSENSE"},
		{"a part with no key", "FREQ=DAILY;=5"},
		{"a part with no value", "FREQ=DAILY;INTERVAL="},
		{"an empty part", "FREQ=DAILY;;INTERVAL=2"},
		{"a duplicated key", "FREQ=DAILY;INTERVAL=2;INTERVAL=3"},
		{"a duplicated BYDAY value", "FREQ=WEEKLY;BYDAY=MO,MO"},
		{"a duplicated BYMONTHDAY value", "FREQ=MONTHLY;BYMONTHDAY=1,1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := domain.ParseRecurrence(tt.rule); !errors.Is(err, domain.ErrUnsupportedRecurrence) {
				t.Errorf("ParseRecurrence(%q) = %v, want ErrUnsupportedRecurrence", tt.rule, err)
			}

			// The same rule through the three public entry points must fail the
			// same way: an unsupported rule can never come back as "no
			// occurrences".
			got, err := domain.Occurrences(tt.rule, d(2026, time.January, 1),
				d(2026, time.January, 1), d(2026, time.December, 31))
			if !errors.Is(err, domain.ErrUnsupportedRecurrence) {
				t.Errorf("Occurrences(%q) = %v, %v, want ErrUnsupportedRecurrence", tt.rule, dateStrings(got), err)
			}
			if _, _, err := domain.NextOccurrence(tt.rule, d(2026, time.January, 1), d(2026, time.January, 1)); !errors.Is(err, domain.ErrUnsupportedRecurrence) {
				t.Errorf("NextOccurrence(%q) = %v, want ErrUnsupportedRecurrence", tt.rule, err)
			}
			if _, _, err := domain.PreviousOccurrence(tt.rule, d(2026, time.January, 1), d(2026, time.December, 31)); !errors.Is(err, domain.ErrUnsupportedRecurrence) {
				t.Errorf("PreviousOccurrence(%q) = %v, want ErrUnsupportedRecurrence", tt.rule, err)
			}
		})
	}
}

func TestParseRecurrenceAcceptsTheSupportedSubset(t *testing.T) {
	tests := []struct {
		name string
		rule string
		want domain.Recurrence
	}{
		{"daily", "FREQ=DAILY", domain.Recurrence{Freq: domain.FreqDaily, Interval: 1}},
		{"daily with an interval", "FREQ=DAILY;INTERVAL=4",
			domain.Recurrence{Freq: domain.FreqDaily, Interval: 4}},
		{"weekly", "FREQ=WEEKLY", domain.Recurrence{Freq: domain.FreqWeekly, Interval: 1}},
		{"weekly with byday", "FREQ=WEEKLY;BYDAY=MO,WE,FR",
			domain.Recurrence{Freq: domain.FreqWeekly, Interval: 1,
				ByDay: []time.Weekday{time.Monday, time.Wednesday, time.Friday}}},
		{"monthly", "FREQ=MONTHLY", domain.Recurrence{Freq: domain.FreqMonthly, Interval: 1}},
		{"monthly with bymonthday", "FREQ=MONTHLY;BYMONTHDAY=1,15",
			domain.Recurrence{Freq: domain.FreqMonthly, Interval: 1, ByMonthDay: []int{1, 15}}},
		{"whitespace around parts", " FREQ=WEEKLY ; BYDAY=MO , FR ",
			domain.Recurrence{Freq: domain.FreqWeekly, Interval: 1,
				ByDay: []time.Weekday{time.Monday, time.Friday}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ParseRecurrence(tt.rule)
			if err != nil {
				t.Fatalf("ParseRecurrence(%q) = %v", tt.rule, err)
			}
			if got.Freq != tt.want.Freq || got.Interval != tt.want.Interval {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
			if len(got.ByDay) != len(tt.want.ByDay) {
				t.Fatalf("ByDay = %v, want %v", got.ByDay, tt.want.ByDay)
			}
			for i := range got.ByDay {
				if got.ByDay[i] != tt.want.ByDay[i] {
					t.Errorf("ByDay[%d] = %v, want %v", i, got.ByDay[i], tt.want.ByDay[i])
				}
			}
			if len(got.ByMonthDay) != len(tt.want.ByMonthDay) {
				t.Fatalf("ByMonthDay = %v, want %v", got.ByMonthDay, tt.want.ByMonthDay)
			}
			for i := range got.ByMonthDay {
				if got.ByMonthDay[i] != tt.want.ByMonthDay[i] {
					t.Errorf("ByMonthDay[%d] = %d, want %d", i, got.ByMonthDay[i], tt.want.ByMonthDay[i])
				}
			}
		})
	}
}

// Matches is the primitive the expansion and both walks are built on, and it is
// exported, so it is tested directly — including the dates before DTSTART that
// the expansion never asks it about because it clamps the window first.
func TestRecurrenceMatches(t *testing.T) {
	tests := []struct {
		name    string
		rule    string
		dtstart domain.Date
		date    domain.Date
		want    bool
	}{
		{"daily on DTSTART", "FREQ=DAILY", d(2026, time.September, 18), d(2026, time.September, 18), true},
		{"daily the day before DTSTART", "FREQ=DAILY", d(2026, time.September, 18), d(2026, time.September, 17), false},
		{"daily years before DTSTART", "FREQ=DAILY", d(2026, time.September, 18), d(1999, time.January, 1), false},
		{"every other day, on", "FREQ=DAILY;INTERVAL=2", d(2026, time.September, 18), d(2026, time.September, 20), true},
		{"every other day, off", "FREQ=DAILY;INTERVAL=2", d(2026, time.September, 18), d(2026, time.September, 21), false},
		{"weekly before DTSTART", "FREQ=WEEKLY", d(2026, time.September, 18), d(2026, time.September, 11), false},
		{"weekly on the same weekday", "FREQ=WEEKLY", d(2026, time.September, 18), d(2026, time.September, 25), true},
		{"weekly on a different weekday", "FREQ=WEEKLY", d(2026, time.September, 18), d(2026, time.September, 24), false},
		{"byday on a listed day", "FREQ=WEEKLY;BYDAY=MO,FR", d(2026, time.September, 14), d(2026, time.September, 18), true},
		{"byday on an unlisted day", "FREQ=WEEKLY;BYDAY=MO,FR", d(2026, time.September, 14), d(2026, time.September, 17), false},
		{"byday in an off week", "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO", d(2026, time.September, 14), d(2026, time.September, 21), false},
		// DTSTART is a Friday; the following Monday is three days later but a
		// week later, so an every-other-week rule skips it.
		{"a byday three days after a friday DTSTART is in the next week", "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,FR",
			d(2026, time.September, 18), d(2026, time.September, 21), false},
		{"and the Monday a week after that is back on", "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,FR",
			d(2026, time.September, 18), d(2026, time.September, 28), true},
		{"monthly before DTSTART", "FREQ=MONTHLY", d(2026, time.September, 18), d(2026, time.August, 18), false},
		{"monthly on the same day of the month", "FREQ=MONTHLY", d(2026, time.September, 18), d(2026, time.October, 18), true},
		{"monthly on a different day", "FREQ=MONTHLY", d(2026, time.September, 18), d(2026, time.October, 19), false},
		{"bymonthday in an off month", "FREQ=MONTHLY;INTERVAL=2;BYMONTHDAY=1", d(2026, time.September, 1), d(2026, time.October, 1), false},
		{"bymonthday in an on month", "FREQ=MONTHLY;INTERVAL=2;BYMONTHDAY=1", d(2026, time.September, 1), d(2026, time.November, 1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := domain.ParseRecurrence(tt.rule)
			if err != nil {
				t.Fatalf("ParseRecurrence(%q) = %v", tt.rule, err)
			}
			if got := r.Matches(tt.dtstart, tt.date); got != tt.want {
				t.Errorf("%q.Matches(%v, %v) = %v, want %v", tt.rule, tt.dtstart, tt.date, got, tt.want)
			}
		})
	}
}

// A zero-value Recurrence — one that never went through the parser — matches
// nothing rather than matching everything.
func TestZeroRecurrenceMatchesNothing(t *testing.T) {
	var r domain.Recurrence
	for i := range 40 {
		day := d(2026, time.September, 1).AddDays(i)
		if r.Matches(d(2026, time.September, 1), day) {
			t.Fatalf("the zero Recurrence matched %v", day)
		}
	}
}

// An open-ended rule over a bounded window returns a bounded result, and a big
// window terminates rather than running away.
func TestOccurrencesOverALargeWindowTerminates(t *testing.T) {
	from := d(1970, time.January, 1)
	to := d(2070, time.January, 1)

	got, err := domain.Occurrences("FREQ=DAILY", from, from, to)
	if err != nil {
		t.Fatalf("Occurrences() = %v", err)
	}
	// A century of days, inclusive at both ends.
	if want := from.DaysUntil(to) + 1; len(got) != want {
		t.Errorf("len = %d, want %d", len(got), want)
	}
	if got[0] != from || got[len(got)-1] != to {
		t.Errorf("window = %v..%v, want %v..%v", got[0], got[len(got)-1], from, to)
	}
}

func TestOccurrencesRefusesAnAbsurdWindow(t *testing.T) {
	from := d(1500, time.January, 1)
	to := d(2500, time.January, 1)

	got, err := domain.Occurrences("FREQ=DAILY", from, from, to)
	if !errors.Is(err, domain.ErrWindowTooLarge) {
		t.Fatalf("Occurrences() = %d dates, %v, want ErrWindowTooLarge", len(got), err)
	}
}

func TestNextOccurrence(t *testing.T) {
	tests := []struct {
		name    string
		rule    string
		dtstart domain.Date
		after   domain.Date
		want    string
		wantOK  bool
	}{
		{"daily", "FREQ=DAILY", d(2026, time.September, 1), d(2026, time.September, 17),
			"2026-09-18", true},
		{"strictly after, so an occurrence today is skipped", "FREQ=DAILY",
			d(2026, time.September, 1), d(2026, time.September, 18), "2026-09-19", true},
		{"weekly byday", "FREQ=WEEKLY;BYDAY=MO,WE,FR", d(2026, time.September, 14),
			d(2026, time.September, 16), "2026-09-18", true},
		{"before DTSTART returns DTSTART itself", "FREQ=WEEKLY", d(2026, time.September, 18),
			d(2020, time.January, 1), "2026-09-18", true},
		{"monthly across a year boundary", "FREQ=MONTHLY;BYMONTHDAY=1",
			d(2026, time.September, 1), d(2026, time.December, 2), "2027-01-01", true},
		{"a leap day is found four years out", "FREQ=MONTHLY;BYMONTHDAY=29",
			d(2026, time.February, 1), d(2026, time.February, 1), "2026-03-29", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := domain.NextOccurrence(tt.rule, tt.dtstart, tt.after)
			if err != nil {
				t.Fatalf("NextOccurrence() = %v", err)
			}
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got.String() != tt.want {
				t.Errorf("NextOccurrence() = %v, want %v", got, tt.want)
			}
		})
	}

	// A rule that can never fire: the 31st of February, once a year. The search
	// is bounded, so this returns rather than looping.
	t.Run("a rule that never fires returns not-found instead of hanging", func(t *testing.T) {
		got, ok, err := domain.NextOccurrence("FREQ=MONTHLY;BYMONTHDAY=31;INTERVAL=12",
			d(2026, time.February, 1), d(2026, time.February, 1))
		if err != nil {
			t.Fatalf("NextOccurrence() = %v", err)
		}
		if ok {
			t.Errorf("NextOccurrence() = %v, true; want not found", got)
		}
	})
}

func TestPreviousOccurrence(t *testing.T) {
	tests := []struct {
		name    string
		rule    string
		dtstart domain.Date
		before  domain.Date
		want    string
		wantOK  bool
	}{
		{"daily", "FREQ=DAILY", d(2026, time.September, 1), d(2026, time.September, 18),
			"2026-09-17", true},
		{"weekly byday skips back over the weekend", "FREQ=WEEKLY;BYDAY=MO,WE,FR",
			d(2026, time.September, 14), d(2026, time.September, 21), "2026-09-18", true},
		{"monthly across a year boundary", "FREQ=MONTHLY;BYMONTHDAY=1",
			d(2025, time.September, 1), d(2026, time.January, 1), "2025-12-01", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := domain.PreviousOccurrence(tt.rule, tt.dtstart, tt.before)
			if err != nil {
				t.Fatalf("PreviousOccurrence() = %v", err)
			}
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got.String() != tt.want {
				t.Errorf("PreviousOccurrence() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("nothing before DTSTART", func(t *testing.T) {
		got, ok, err := domain.PreviousOccurrence("FREQ=DAILY",
			d(2026, time.September, 18), d(2026, time.September, 18))
		if err != nil {
			t.Fatalf("PreviousOccurrence() = %v", err)
		}
		if ok {
			t.Errorf("PreviousOccurrence() = %v, true; want not found", got)
		}
	})

	t.Run("a rule that never fired returns not-found instead of hanging", func(t *testing.T) {
		// The search back is capped at maxSearchDays, which this DTSTART is
		// well beyond.
		got, ok, err := domain.PreviousOccurrence("FREQ=DAILY",
			d(1500, time.January, 1), d(2026, time.September, 18))
		if err != nil {
			t.Fatalf("PreviousOccurrence() = %v", err)
		}
		if !ok {
			t.Fatalf("PreviousOccurrence() = %v, false; want the day before", got)
		}
		if got.String() != "2026-09-17" {
			t.Errorf("PreviousOccurrence() = %v, want 2026-09-17", got)
		}
	})

	t.Run("a sparse rule outside the search horizon is not found", func(t *testing.T) {
		// The only occurrences are in 1500; the horizon back from 2026 is a
		// century, so the answer is "none", bounded, not a hang.
		got, ok, err := domain.PreviousOccurrence("FREQ=MONTHLY;BYMONTHDAY=31;INTERVAL=12",
			d(1500, time.February, 1), d(2026, time.September, 18))
		if err != nil {
			t.Fatalf("PreviousOccurrence() = %v", err)
		}
		if ok {
			t.Errorf("PreviousOccurrence() = %v, true; want not found", got)
		}
	})
}

// Expansion is a pure function of the rule and the dates: same input, same
// output, every time.
func TestOccurrencesAreDeterministic(t *testing.T) {
	first, err := domain.Occurrences("FREQ=WEEKLY;BYDAY=MO,WE,FR",
		d(2026, time.September, 14), d(2026, time.September, 1), d(2026, time.December, 31))
	if err != nil {
		t.Fatalf("Occurrences() = %v", err)
	}
	for range 3 {
		again, err := domain.Occurrences("FREQ=WEEKLY;BYDAY=MO,WE,FR",
			d(2026, time.September, 14), d(2026, time.September, 1), d(2026, time.December, 31))
		if err != nil {
			t.Fatalf("Occurrences() = %v", err)
		}
		if !equalStrings(dateStrings(first), dateStrings(again)) {
			t.Fatalf("expansion is not deterministic")
		}
	}
	// And it is ascending with no duplicates.
	for i := 1; i < len(first); i++ {
		if !first[i].After(first[i-1]) {
			t.Fatalf("occurrences are not strictly ascending at %d: %v then %v", i, first[i-1], first[i])
		}
	}
}

// Expand and Next agree: walking with Next reproduces the expansion exactly.
func TestNextWalkAgreesWithExpand(t *testing.T) {
	rules := []string{
		"FREQ=DAILY;INTERVAL=2",
		"FREQ=WEEKLY;BYDAY=TU,TH",
		"FREQ=WEEKLY;INTERVAL=3",
		"FREQ=MONTHLY;BYMONTHDAY=5,20",
	}
	dtstart := d(2026, time.January, 1)
	to := d(2026, time.June, 30)

	for _, rule := range rules {
		t.Run(rule, func(t *testing.T) {
			expanded, err := domain.Occurrences(rule, dtstart, dtstart, to)
			if err != nil {
				t.Fatalf("Occurrences() = %v", err)
			}

			var walked []domain.Date
			cursor := dtstart.AddDays(-1)
			for {
				next, ok, err := domain.NextOccurrence(rule, dtstart, cursor)
				if err != nil {
					t.Fatalf("NextOccurrence() = %v", err)
				}
				if !ok || next.After(to) {
					break
				}
				walked = append(walked, next)
				cursor = next
			}
			if !equalStrings(dateStrings(walked), dateStrings(expanded)) {
				t.Errorf("walking with Next gave %v, expansion gave %v",
					dateStrings(walked), dateStrings(expanded))
			}

			// And walking backwards gives the same set in reverse.
			var back []domain.Date
			cursor = to.AddDays(1)
			for {
				prev, ok, err := domain.PreviousOccurrence(rule, dtstart, cursor)
				if err != nil {
					t.Fatalf("PreviousOccurrence() = %v", err)
				}
				if !ok {
					break
				}
				back = append(back, prev)
				cursor = prev
			}
			reversed := make([]domain.Date, len(back))
			for i, x := range back {
				reversed[len(back)-1-i] = x
			}
			if !equalStrings(dateStrings(reversed), dateStrings(expanded)) {
				t.Errorf("walking with Previous gave %v, expansion gave %v",
					dateStrings(reversed), dateStrings(expanded))
			}
		})
	}
}
