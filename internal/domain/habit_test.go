package domain_test

import (
	"testing"
	"time"

	"nexus/internal/domain"
)

func TestHabitCheckValidate(t *testing.T) {
	valid := domain.HabitCheck{NodeID: "h1", Date: domain.NewDate(2026, time.September, 18)}

	t.Run("a valid check", func(t *testing.T) {
		if err := valid.Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	})

	bad := []struct {
		name  string
		field string
		mut   func(*domain.HabitCheck)
	}{
		{"empty node id", "node_id", func(c *domain.HabitCheck) { c.NodeID = "" }},
		{"blank node id", "node_id", func(c *domain.HabitCheck) { c.NodeID = "\t" }},
		{"missing date", "date", func(c *domain.HabitCheck) { c.Date = domain.Date{} }},
		{"impossible date", "date", func(c *domain.HabitCheck) {
			c.Date = domain.NewDate(2026, time.February, 29)
		}},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			c := valid
			tt.mut(&c)
			assertValidationError(t, c.Validate(), "habit_check", tt.field)
		})
	}
}

// habit_checks is keyed by (node_id, date), so a check is a value: two checks
// for the same habit on the same day are the same fact, and must compare equal.
func TestHabitCheckIsComparableByValue(t *testing.T) {
	a := domain.HabitCheck{NodeID: "h1", Date: domain.NewDate(2026, time.September, 18)}
	b := domain.HabitCheck{NodeID: "h1", Date: domain.NewDate(2026, time.September, 18)}
	c := domain.HabitCheck{NodeID: "h1", Date: domain.NewDate(2026, time.September, 19)}

	if a != b {
		t.Errorf("%v != %v, want the same day on the same habit to be one fact", a, b)
	}
	if a == c {
		t.Errorf("%v == %v, want different days to be different facts", a, c)
	}
}
