# Project Planner

Fit tasks into your Google Calendar schedule automatically.

## Features

- Parse project tasks from a YAML file
- Schedule tasks based on your availability
- Integrate with Google Calendar to avoid conflicts
- Support for priorities, due dates, and tags
- Configurable work hours and days

## Installation

### Prerequisites

- Go 1.21 or higher
- Google Cloud project with Calendar API enabled
- Google OAuth 2.0 credentials

### Build

```bash
# Clone the repository
git clone https://github.com/KelvinMcClean/project-planner.git
cd project-planner

# Install dependencies
go mod download

# Build the application
go build ./cmd/project-planner
```

## Setup Google Calendar API

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the "Google Calendar API"
4. Create OAuth 2.0 credentials (OAuth client ID)
5. Download the credentials JSON file and save it as `credentials.json` in the project root or in `pkg/calendar/`
6. Add your email to the authorized redirect URIs in the OAuth consent screen

### Required Scopes

- `https://www.googleapis.com/auth/calendar`
- `https://www.googleapis.com/auth/calendar.events`

## Usage

### Create a Project File

Create a YAML file with your project tasks. See `examples/project.yml` for a complete example.

```yaml
name: My Project
description: Project description
start_date: 2024-01-01
end_date: 2024-01-31

settings:
  work_hours: "09:00-17:00"
  work_days: ["mon", "tue", "wed", "thu", "fri"]
  buffer_time: "15m"
  max_tasks_per_day: 5

tasks:
  - name: Task 1
    description: Task description
    duration: 2h
    priority: 5
    due_date: 2024-01-10
    tags: ["important", "urgent"]
  - name: Task 2
    duration: 1h30m
    priority: 3
```

### Run the Planner

```bash
# Preview the schedule (dry run)
./project-planner --file myproject.yml --dry-run --verbose

# Actually add events to your calendar
./project-planner --file myproject.yml --verbose
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--file` | Path to the YAML project file | `project.yml` |
| `--dry-run` | Preview the schedule without adding to calendar | `false` |
| `--verbose` | Show detailed output | `false` |

## YAML File Format

### Project Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Project name |
| `description` | string | No | Project description |
| `start_date` | date | Yes | Project start date (YYYY-MM-DD) |
| `end_date` | date | Yes | Project end date (YYYY-MM-DD) |
| `tasks` | array | Yes | List of tasks |
| `settings` | object | No | Scheduling settings |

### Task Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Task name |
| `description` | string | No | Task description |
| `duration` | string | Yes | Duration (e.g., "2h", "30m", "1h30m") |
| `priority` | int | No | Priority (higher = more important) |
| `due_date` | date | No | Due date (YYYY-MM-DD) |
| `tags` | array | No | List of tags |

### Settings Structure

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `work_hours` | string | "09:00-17:00" | Daily working hours (HH:MM-HH:MM) |
| `work_days` | array | ["mon", "tue", "wed", "thu", "fri"] | Days of week to schedule on |
| `buffer_time` | string | "0" | Buffer time between tasks (e.g., "15m") |
| `max_tasks_per_day` | int | 10 | Maximum tasks to schedule per day |
| `avoid_overlap` | bool | true | Avoid overlapping with existing calendar events |

### Supported Day Names

- `mon` - Monday
- `tue` - Tuesday
- `wed` - Wednesday
- `thu` - Thursday
- `fri` - Friday
- `sat` - Saturday
- `sun` - Sunday

### Duration Format

- `30m` - 30 minutes
- `1h` - 1 hour
- `1h30m` - 1 hour and 30 minutes
- `2h` - 2 hours
- `1d` - 1 day (24 hours)

## How It Works

1. **Parse**: The application reads your YAML file and parses the project definition
2. **Validate**: Checks for required fields and valid data
3. **Fetch Calendar**: Retrieves existing events from your Google Calendar
4. **Schedule**: Algorithms determine the best time slots for each task
5. **Create Events**: Adds scheduled tasks as events to your Google Calendar

### Scheduling Algorithm

Tasks are scheduled based on:
1. Priority (higher priority tasks first)
2. Due date (tasks with earlier due dates first)
3. Available time slots (based on your work hours and existing events)
4. Buffer time between tasks

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
