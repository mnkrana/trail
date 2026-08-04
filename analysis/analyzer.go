package analysis

import (
	"sort"
	"time"

	"google.golang.org/api/calendar/v3"
)

type WeekAnalysis struct {
	Events          []*calendar.Event
	TotalEvents     int
	TotalMinutes    int
	AverageDuration float64
	ByCategory      map[string]int
	BusyDays        []DaySummary
}

type DaySummary struct {
	Date         time.Time
	EventCount   int
	TotalMinutes int
}

// Category color mapping (ANSI escape codes)
var categoryColors = map[string]string{
	"Meeting":  "\033[34m", // Blue
	"Focus":    "\033[32m", // Green
	"Personal": "\033[33m", // Yellow
	"Health":   "\033[35m", // Magenta
	"Learning": "\033[36m", // Cyan
	"Admin":    "\033[37m", // White
	"Social":   "\033[91m", // Light Red
	"Other":    "\033[90m", // Gray
}

func AnalyzeEvents(events []*calendar.Event) *WeekAnalysis {
	analysis := &WeekAnalysis{
		Events:      events,
		TotalEvents: len(events),
		ByCategory:  make(map[string]int),
	}

	dayMap := make(map[string]*DaySummary)

	for _, event := range events {
		// Parse start time
		startStr := event.Start.DateTime
		if startStr == "" {
			startStr = event.Start.Date
		}

		endStr := event.End.DateTime
		if endStr == "" {
			endStr = event.End.Date
		}

		startTime, _ := time.Parse(time.RFC3339, startStr)
		endTime, _ := time.Parse(time.RFC3339, endStr)

		duration := int(endTime.Sub(startTime).Minutes())
		analysis.TotalMinutes += duration

		// Categorize event
		category := GetCategoryFromEvent(event)
		analysis.ByCategory[category] += duration

		// Track daily summary
		dayKey := startTime.Format("2006-01-02")
		if dayMap[dayKey] == nil {
			dayMap[dayKey] = &DaySummary{
				Date: startTime,
			}
		}
		dayMap[dayKey].EventCount++
		dayMap[dayKey].TotalMinutes += duration
	}

	// Calculate average duration
	if analysis.TotalEvents > 0 {
		analysis.AverageDuration = float64(analysis.TotalMinutes) / float64(analysis.TotalEvents)
	}

	// Convert day map to sorted slice
	for _, day := range dayMap {
		analysis.BusyDays = append(analysis.BusyDays, *day)
	}

	sort.Slice(analysis.BusyDays, func(i, j int) bool {
		return analysis.BusyDays[i].Date.Before(analysis.BusyDays[j].Date)
	})

	return analysis
}

func GetCategoryFromEvent(event *calendar.Event) string {
	summary := event.Summary
	description := event.Description

	// Simple keyword-based categorization
	// You can enhance this with more sophisticated logic
	keywords := map[string][]string{
		"Meeting":  {"meeting", "standup", "sync", "call", "1:1", "review", "demo"},
		"Focus":    {"focus", "deep work", "coding", "development", "implementation", "debugging"},
		"Personal": {"personal", "appointment", "errands", "family"},
		"Health":   {"gym", "workout", "exercise", "doctor", "health", "meditation", "yoga"},
		"Learning": {"learning", "course", "study", "reading", "tutorial", "documentation"},
		"Admin":    {"admin", "planning", "retrospective", "backlog", "sprint"},
		"Social":   {"social", "lunch", "coffee", "team building", "happy hour"},
	}

	// Check calendar color if available (Google Calendar API provides colorId)
	if event.ColorId != "" {
		// Map Google Calendar color IDs to categories
		// Color IDs: 1=Lavender, 2=Sage, 3=Grape, 4=Flamingo, 5=Banana,
		//            6=Tangerine, 7=Peacock, 8=Graphite, 9=Blueberry, 10=Basil, 11=Tomato
		colorMap := map[string]string{
			"1":  "Admin",    // Lavender
			"2":  "Health",   // Sage
			"3":  "Meeting",  // Grape
			"4":  "Personal", // Flamingo
			"5":  "Learning", // Banana
			"6":  "Social",   // Tangerine
			"7":  "Meeting",  // Peacock
			"8":  "Other",    // Graphite
			"9":  "Focus",    // Blueberry
			"10": "Health",   // Basil
			"11": "Other",    // Tomato
		}
		if category, ok := colorMap[event.ColorId]; ok {
			return category
		}
	}

	// Fallback to keyword matching
	combined := summary + " " + description
	for category, words := range keywords {
		for _, word := range words {
			if containsIgnoreCase(combined, word) {
				return category
			}
		}
	}

	return "Other"
}

func GetCategoryColor(category string) string {
	if color, ok := categoryColors[category]; ok {
		return color
	}
	return categoryColors["Other"]
}

func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
