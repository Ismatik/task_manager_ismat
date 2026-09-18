package service_test

import (
	"testing"
	"time"

	"nexus/internal/service"
)

func TestFixedClockIsDeterministic(t *testing.T) {
	// 2026-09-18 is a Friday: the column<->due edge case that Stage 1 has to
	// get right, and the reason a fake clock exists at all.
	at := time.Date(2026, time.September, 18, 9, 30, 0, 0, time.UTC)

	clock := service.FixedClock(at)

	if got := clock(); !got.Equal(at) {
		t.Errorf("FixedClock()() = %v, want %v", got, at)
	}
	if got := clock(); !got.Equal(at) {
		t.Errorf("FixedClock() is not stable across calls: got %v, want %v", got, at)
	}
	if got := clock().Weekday(); got != time.Friday {
		t.Errorf("Weekday() = %v, want Friday", got)
	}
}

func TestSystemClockSatisfiesClock(t *testing.T) {
	var clock service.Clock = service.SystemClock

	before := time.Now()
	got := clock()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("SystemClock() = %v, want a time within [%v, %v]", got, before, after)
	}
}
