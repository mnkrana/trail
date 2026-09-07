# Trail - Calendar Analytics CLI

A simple CLI tool to analyze your Google Calendar events for the last week. Track your progress, see time distribution across categories, and get insights into how you spend your time.

## Features

- 📅 Analyze last week's calendar events (Monday to Sunday)
- 🎨 Color-coded categories (Meetings, Focus, Personal, Health, etc.)
- 📊 Time breakdown by category
- 📈 Daily summary with event counts
- 🔐 Secure OAuth2 authentication with token persistence
- 🚀 Install directly with `go install` — no cloning needed

## Install

Requires Go 1.23+.

```bash
go install github.com/mnkrana/trail@latest
```

This puts the `trail` binary in `$GOPATH/bin`, so make sure that's on your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Check the version:

```bash
trail --version
```

## Setup

### 1. Google Cloud Console Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Calendar API
4. Create OAuth 2.0 credentials:
   - Go to APIs & Services → Credentials
   - Click "Create Credentials" → "OAuth client ID"
   - Application type: "Desktop app" (or "Web application" with redirect URI)
   - Add redirect URI: `http://localhost:8080/api/oauth/calendar/callback`
5. Note the Client ID and Client Secret

### 2. Configure Credentials (one-time)

```bash
trail configure
```

You'll be prompted for your OAuth Client ID and Client Secret, which are saved to
`~/.config/trail/config.json` (with `0600` permissions).

Alternatively, you can set environment variables instead of using `trail configure`:

```bash
export OAUTH_CLIENT_ID="your_client_id_here"
export OAUTH_CLIENT_SECRET="your_client_secret_here"
```

> If you set the environment variables, they take precedence over the config file.
> You can also put them in a `.env` file in the directory you run `trail` from.

### 3. Authenticate (one-time)

```bash
trail auth
```

This starts a local server on `localhost:8080`, opens your browser to the Google consent screen, captures the code automatically, and saves tokens to `~/.config/trail/tokens.json`.

### 4. Run

```bash
# Last week (Monday to Sunday)
trail run --last-week

# Specific date range
trail run --start 2024-01-15 --end 2024-01-21
```

Future runs are silent — no auth prompt. If you haven't authenticated, it tells you to run `trail auth` first.

## Categories

Events are automatically categorized based on keywords and calendar colors:

- 🔵 **Meeting** - Meetings, standups, calls, reviews
- 🟢 **Focus** - Deep work, coding, development
- 🟡 **Personal** - Personal appointments, errands
- 🟣 **Health** - Gym, workouts, meditation
- 🔷 **Learning** - Courses, reading, documentation
- ⚪ **Admin** - Planning, retrospectives
- 🔴 **Social** - Team events, coffee chats
- ⬜ **Other** - Uncategorized events

## Output

The tool displays:

1. **Events Table** - All events with time, name, duration, and category
2. **Summary Statistics**:
   - Total events and time
   - Average event duration
   - Time breakdown by category with percentages
   - Daily summary with event counts

## Example Output

```
Analyzing calendar from Jan 15 to Jan 21, 2024

┌───────────┬──────────────────────────┬──────────┬──────────┐
│ TIME      │ EVENT                    │ DURATION │ CATEGORY │
├───────────┼──────────────────────────┼──────────┼──────────┤
│ Mon 09:00 │ Team Standup             │ 30 min   │ Meeting  │
│ Mon 10:00 │ Deep Work Session        │ 120 min  │ Focus    │
│ Mon 14:00 │ Lunch with Team          │ 60 min   │ Social   │
│ ...       │ ...                      │ ...      │ ...      │
└───────────┴──────────────────────────┴──────────┴──────────┘

📊 Summary
─────────────────────────────────────────
Total Events: 25
Total Time: 45 hours 30 minutes
Average Duration: 109.2 minutes

📁 Time by Category:
  Meeting     : 12h 30m (27.5%)
  Focus       : 18h 00m (39.6%)
  Personal    :  3h 00m (6.6%)
  Health      :  4h 00m (8.8%)
  Social      :  5h 00m (11.0%)
  Other       :  3h 00m (6.6%)

📅 Busiest Days:
  Mon Jan 15: 6 events, 8 hours
  Wed Jan 17: 5 events, 7 hours
  Fri Jan 19: 7 events, 9 hours
```

## Development

```bash
# Run tests
go test ./...

# Build for current platform
go build -o trail .

# Build for other platforms
GOOS=linux GOARCH=amd64 go build -o trail-linux .
GOOS=darwin GOARCH=arm64 go build -o trail-mac .
```

## License

MIT
