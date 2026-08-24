package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/joho/godotenv"
	"github.com/mnkrana/trail/analysis"
	"github.com/mnkrana/trail/calendar"
	"github.com/spf13/cobra"
)

func main() {
	// Load .env file if present
	godotenv.Load()

	var rootCmd = &cobra.Command{
		Use:   "trail",
		Short: "Calendar analytics CLI - track your weekly progress",
	}

	var authCmd = &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Google Calendar (one-time)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := calendar.NewOAuthConfig()
			return calendar.RunAuthServer(cmd.Context(), cfg)
		},
	}

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Analyze calendar events",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			startDate, _ := cmd.Flags().GetString("start")
			endDate, _ := cmd.Flags().GetString("end")
			lastWeek, _ := cmd.Flags().GetBool("last-week")
			from, to, err := resolveRange(lastWeek, startDate, endDate, time.Now())
			if err != nil {
				return err
			}

			fmt.Printf("Analyzing calendar from %s to %s\n\n", from.Format("Jan 2"), to.Format("Jan 2, 2006"))

			// Initialize calendar client
			calClient, err := calendar.NewClient(ctx)
			if err != nil {
				return err
			}

			// Fetch events
			events, err := calClient.FetchEvents(ctx, from, to)
			if err != nil {
				return fmt.Errorf("failed to fetch events: %v", err)
			}

			if len(events) == 0 {
				fmt.Println("No events found for this period.")
				return nil
			}

			// Analyze events
			analysis := analysis.AnalyzeEvents(events)

			// Display results
			displayResults(analysis)

			return nil
		},
	}

	runCmd.Flags().Bool("last-week", false, "Analyze last week (Monday to Sunday)")
	runCmd.Flags().String("start", "", "Start date (YYYY-MM-DD)")
	runCmd.Flags().String("end", "", "End date (YYYY-MM-DD)")

	rootCmd.AddCommand(authCmd, runCmd)
	rootCmd.SetArgs(os.Args[1:])

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// resolveRange computes the analysis window. With lastWeek set, it returns the
// previous Monday-to-Sunday week relative to now; otherwise it parses the
// explicit start/end dates (end extended to end-of-day).
func resolveRange(lastWeek bool, startDate, endDate string, now time.Time) (time.Time, time.Time, error) {
	if lastWeek {
		daysSinceMonday := int(now.Weekday()) - 1
		if daysSinceMonday < 0 {
			daysSinceMonday = 6
		}
		lastMonday := now.AddDate(0, 0, -daysSinceMonday-7)
		lastMonday = time.Date(lastMonday.Year(), lastMonday.Month(), lastMonday.Day(), 0, 0, 0, 0, lastMonday.Location())
		lastSunday := lastMonday.AddDate(0, 0, 6)
		lastSunday = time.Date(lastSunday.Year(), lastSunday.Month(), lastSunday.Day(), 23, 59, 59, 0, lastSunday.Location())
		return lastMonday, lastSunday, nil
	}

	if startDate == "" || endDate == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("please specify --last-week or both --start and --end dates")
	}

	from, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format: %v", err)
	}
	to, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format: %v", err)
	}
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, to.Location())
	return from, to, nil
}

func displayResults(a *analysis.WeekAnalysis) {
	// Create main events table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Set header
	t.AppendHeader(table.Row{"Time", "Event", "Duration", "Category"})

	// Set column styles
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:     "Event",
			WidthMin: 30,
			WidthMax: 50,
		},
		{
			Name:     "Category",
			WidthMin: 15,
		},
	})

	// Add events
	for _, event := range a.Events {
		category := analysis.GetCategoryFromEvent(event)
		color := analysis.GetCategoryColor(category)

		// Parse start time
		startStr := event.Start.DateTime
		if startStr == "" {
			startStr = event.Start.Date
		}
		startTime, _ := time.Parse(time.RFC3339, startStr)

		// Calculate duration
		endStr := event.End.DateTime
		if endStr == "" {
			endStr = event.End.Date
		}
		endTime, _ := time.Parse(time.RFC3339, endStr)
		duration := int(endTime.Sub(startTime).Minutes())

		t.AppendRow(table.Row{
			startTime.Format("Mon 15:04"),
			event.Summary,
			fmt.Sprintf("%d min", duration),
			color + category + "\033[0m",
		})
	}

	// Set style
	t.SetStyle(table.StyleLight)

	t.Render()

	// Display summary
	fmt.Printf("\n📊 Summary\n")
	fmt.Printf("─────────────────────────────────────────\n")
	fmt.Printf("Total Events: %d\n", a.TotalEvents)
	fmt.Printf("Total Time: %d hours %d minutes\n", a.TotalMinutes/60, a.TotalMinutes%60)
	fmt.Printf("Average Duration: %.1f minutes\n", a.AverageDuration)

	if len(a.ByCategory) > 0 {
		fmt.Printf("\n📁 Time by Category:\n")
		for category, minutes := range a.ByCategory {
			color := analysis.GetCategoryColor(category)
			hours := minutes / 60
			mins := minutes % 60
			percentage := float64(minutes) / float64(a.TotalMinutes) * 100
			fmt.Printf("  %s%-12s\033[0m: %2dh %02dm (%.1f%%)\n", color, category, hours, mins, percentage)
		}
	}

	if len(a.BusyDays) > 0 {
		fmt.Printf("\n📅 Busiest Days:\n")
		for _, day := range a.BusyDays {
			fmt.Printf("  %s: %d events, %d hours\n",
				day.Date.Format("Mon Jan 2"),
				day.EventCount,
				day.TotalMinutes/60)
		}
	}
}
