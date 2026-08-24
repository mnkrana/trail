package analysis

import (
	"testing"
	"time"

	"google.golang.org/api/calendar/v3"
)

func mkEvent(summary, description, colorID string, start, end time.Time) *calendar.Event {
	return &calendar.Event{
		Summary:     summary,
		Description: description,
		ColorId:     colorID,
		Start:       &calendar.EventDateTime{DateTime: start.Format(time.RFC3339)},
		End:         &calendar.EventDateTime{DateTime: end.Format(time.RFC3339)},
	}
}

func TestAnalyzeEventsTotals(t *testing.T) {
	base := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	events := []*calendar.Event{
		mkEvent("Team Standup", "", "", base.Add(9*time.Hour), base.Add(9*time.Hour+30*time.Minute)),
		mkEvent("Gym session", "", "", base.Add(24*time.Hour+7*time.Hour), base.Add(24*time.Hour+8*time.Hour)),
		mkEvent("Coffee chat", "", "", base.Add(48*time.Hour+10*time.Hour), base.Add(48*time.Hour+11*time.Hour)),
	}

	a := AnalyzeEvents(events)

	if a.TotalEvents != 3 {
		t.Errorf("TotalEvents = %d, want 3", a.TotalEvents)
	}
	if a.TotalMinutes != 150 {
		t.Errorf("TotalMinutes = %d, want 150", a.TotalMinutes)
	}
	if a.AverageDuration != 50.0 {
		t.Errorf("AverageDuration = %v, want 50", a.AverageDuration)
	}
	want := map[string]int{"Meeting": 30, "Health": 60, "Social": 60}
	for cat, minutes := range want {
		if got := a.ByCategory[cat]; got != minutes {
			t.Errorf("ByCategory[%q] = %d, want %d", cat, got, minutes)
		}
	}
	if len(a.ByCategory) != len(want) {
		t.Errorf("ByCategory has %d entries, want %d: %v", len(a.ByCategory), len(want), a.ByCategory)
	}
}

func TestAnalyzeEventsBusyDaysSorted(t *testing.T) {
	base := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	events := []*calendar.Event{
		mkEvent("Standup", "", "", base.Add(48*time.Hour), base.Add(48*time.Hour+30*time.Minute)),
		mkEvent("Gym", "", "", base.Add(24*time.Hour), base.Add(24*time.Hour+60*time.Minute)),
		mkEvent("Morning coffee", "", "", base.Add(9*time.Hour), base.Add(9*time.Hour+60*time.Minute)),
		mkEvent("Afternoon coffee", "", "", base.Add(14*time.Hour), base.Add(14*time.Hour+30*time.Minute)),
	}

	a := AnalyzeEvents(events)

	if len(a.BusyDays) != 3 {
		t.Fatalf("len(BusyDays) = %d, want 3", len(a.BusyDays))
	}
	for i := 1; i < len(a.BusyDays); i++ {
		if a.BusyDays[i].Date.Before(a.BusyDays[i-1].Date) {
			t.Errorf("BusyDays not sorted: index %d (%v) before %d (%v)",
				i, a.BusyDays[i].Date, i-1, a.BusyDays[i-1].Date)
		}
	}

	first := a.BusyDays[0]
	if first.EventCount != 2 || first.TotalMinutes != 90 {
		t.Errorf("first day = %+v, want 2 events / 90 min (two events on same day merged)", first)
	}
}

func TestAnalyzeEventsEmpty(t *testing.T) {
	a := AnalyzeEvents(nil)

	if a.TotalEvents != 0 || a.TotalMinutes != 0 || a.AverageDuration != 0 {
		t.Errorf("expected zeroed analysis, got %+v", a)
	}
	if a.ByCategory == nil {
		t.Error("ByCategory should be an initialized empty map")
	}
	if len(a.BusyDays) != 0 {
		t.Errorf("BusyDays should be empty, got %v", a.BusyDays)
	}
}

// Note: the keyword map is iterated in random order, so inputs below are
// chosen to match exactly one category to keep results deterministic.
func TestGetCategoryFromEventKeywords(t *testing.T) {
	tests := []struct {
		name        string
		summary     string
		description string
		colorID     string
		want        string
	}{
		{name: "keyword in summary", summary: "Team Standup", want: "Meeting"},
		{name: "case-insensitive keyword", summary: "DEEP WORK BLOCK", want: "Focus"},
		{name: "keyword only in description", summary: "Untitled event", description: "morning yoga class", want: "Health"},
		{name: "no match falls back to Other", summary: "Random Event", description: "", want: "Other"},
		{name: "color ID takes precedence over keywords", summary: "gym session", colorID: "3", want: "Meeting"},
		{name: "unknown color ID falls back to keywords", summary: "standup", colorID: "99", want: "Meeting"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &calendar.Event{
				Summary:     tt.summary,
				Description: tt.description,
				ColorId:     tt.colorID,
			}
			if got := GetCategoryFromEvent(event); got != tt.want {
				t.Errorf("GetCategoryFromEvent(%q, %q, colorId=%q) = %q, want %q",
					tt.summary, tt.description, tt.colorID, got, tt.want)
			}
		})
	}
}

func TestGetCategoryColor(t *testing.T) {
	if got := GetCategoryColor("Meeting"); got != "\033[34m" {
		t.Errorf("GetCategoryColor(Meeting) = %q, want blue escape code", got)
	}
	if got := GetCategoryColor("Nonexistent"); got != "\033[90m" {
		t.Errorf("GetCategoryColor(Nonexistent) = %q, want gray fallback", got)
	}
}
