package main

import (
	"strings"
	"testing"
	"time"
)

func TestResolveRangeLastWeek(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		wantFrom time.Time
		wantTo   time.Time
	}{
		{
			name:     "mid-week Wednesday",
			now:      time.Date(2026, 8, 19, 15, 4, 0, 0, time.Local),
			wantFrom: time.Date(2026, 8, 10, 0, 0, 0, 0, time.Local),
			wantTo:   time.Date(2026, 8, 16, 23, 59, 59, 0, time.Local),
		},
		{
			name:     "Monday",
			now:      time.Date(2026, 8, 17, 9, 0, 0, 0, time.Local),
			wantFrom: time.Date(2026, 8, 10, 0, 0, 0, 0, time.Local),
			wantTo:   time.Date(2026, 8, 16, 23, 59, 59, 0, time.Local),
		},
		{
			name:     "Sunday counts as end of current week",
			now:      time.Date(2026, 8, 23, 22, 0, 0, 0, time.Local),
			wantFrom: time.Date(2026, 8, 10, 0, 0, 0, 0, time.Local),
			wantTo:   time.Date(2026, 8, 16, 23, 59, 59, 0, time.Local),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to, err := resolveRange(true, "", "", tt.now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !from.Equal(tt.wantFrom) {
				t.Errorf("from = %v, want %v", from, tt.wantFrom)
			}
			if !to.Equal(tt.wantTo) {
				t.Errorf("to = %v, want %v", to, tt.wantTo)
			}
		})
	}
}

func TestResolveRangeExplicitDates(t *testing.T) {
	from, to, err := resolveRange(false, "2024-01-15", "2024-01-21", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantFrom := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	wantTo := time.Date(2024, 1, 21, 23, 59, 59, 0, time.UTC)
	if !from.Equal(wantFrom) {
		t.Errorf("from = %v, want %v", from, wantFrom)
	}
	if !to.Equal(wantTo) {
		t.Errorf("to = %v, want %v (end should extend to end-of-day)", to, wantTo)
	}
}

func TestResolveRangeErrors(t *testing.T) {
	tests := []struct {
		name      string
		lastWeek  bool
		startDate string
		endDate   string
		wantErr   string
	}{
		{
			name:    "no flags",
			wantErr: "please specify --last-week or both --start and --end dates",
		},
		{
			name:      "only start date",
			startDate: "2024-01-15",
			wantErr:   "please specify --last-week or both --start and --end dates",
		},
		{
			name:      "bad start date format",
			lastWeek:  false,
			startDate: "15-01-2024",
			endDate:   "2024-01-21",
			wantErr:   "invalid start date format",
		},
		{
			name:      "bad end format",
			startDate: "2024-01-15",
			endDate:   "not-a-date",
			wantErr:   "invalid end date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := resolveRange(tt.lastWeek, tt.startDate, tt.endDate, time.Now())
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
