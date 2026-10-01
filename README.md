# Project Planner

Fit tasks into your Google Calendar schedule automatically.

## Features

- Parse project tasks from a YAML file
- Schedule tasks based on your availability
- Integrate with Google Calendar to avoid conflicts
- Support for priorities, due dates, and tags
- **Daily customizable work hours** (different hours for each day)
- **Hours-per-week limit** to control total workload
- **Task dependencies** to ensure proper ordering
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
  # Daily customizable work hours
  work_hours:
    mon: "09:00-17:00"
    tue: "08:00-16:00"
    wed: "09:00-17:00"
    thu: "09:00-17:00"
    fri: "09:00-13:00"
  
  # Or use a single work_hours string for all days:
  # work_hours: "09:00-17:00"
  
  work_days: ["mon", "tue", "wed", "thu", "fri"]
  buffer_time: "15m"
  max_tasks_per_day: 5
  hours_per_week: 40
  avoid_overlap: true

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
    depends_on: ["Task 1"]  # Task 2 depends on Task 1
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
| `depends_on` | array | No | List of task names this task depends on |

### Settings Structure

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `work_hours` | string or map | "09:00-17:00" | Daily working hours. Can be a single string (applied to all days) or a map with day keys ("mon", "tue", etc.) |
| `work_days` | array | ["mon", "tue", "wed", "thu", "fri"] | Days of week to schedule on |
| `buffer_time` | string | "0" | Buffer time between tasks (e.g., "15m") |
| `max_tasks_per_day` | int | 10 | Maximum tasks to schedule per day |
| `hours_per_week` | float | 0 (no limit) | Maximum hours to schedule per week |
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

## Task Dependencies

Tasks can depend on other tasks using the `depends_on` field. The scheduler will ensure that dependent tasks are scheduled after their dependencies.

```yaml
tasks:
  - name: Design Database
    duration: 2h
    
  - name: Implement API
    duration: 4h
    depends_on: ["Design Database"]
    
  - name: Test API
    duration: 2h
    depends_on: ["Implement API"]
```

In this example:
1. "Design Database" will be scheduled first
2. "Implement API" will be scheduled after "Design Database" is complete
3. "Test API" will be scheduled after "Implement API" is complete

### Dependency Resolution

- Dependencies are resolved automatically before scheduling a task
- If a dependency cannot be scheduled, the dependent task will be marked as having dependency issues
- Circular dependencies are not detected automatically (avoid them in your project file)

## Hours Per Week Limit

The `hours_per_week` setting allows you to limit the total hours scheduled per week. This is useful for:

- Controlling workload
- Ensuring work-life balance
- Meeting contractual hour limits

```yaml
settings:
  hours_per_week: 40  # Limit to 40 hours per week
```

## Daily Customizable Work Hours

You can specify different work hours for each day of the week:

```yaml
settings:
  work_hours:
    mon: "09:00-17:00"  # Monday: 9 AM to 5 PM
    tue: "08:00-16:00"  # Tuesday: 8 AM to 4 PM
    wed: "09:00-17:00"
    thu: "09:00-17:00"
    fri: "09:00-13:00"  # Friday: 9 AM to 1 PM (half day)
```

Or use a single string for all days:

```yaml
settings:
  work_hours: "09:00-17:00"  # All days: 9 AM to 5 PM
```

## How It Works

1. **Parse**: The application reads your YAML file and parses the project definition
2. **Validate**: Checks for required fields and valid data
3. **Fetch Calendar**: Retrieves existing events from your Google Calendar
4. **Schedule**: Algorithms determine the best time slots for each task, respecting:
   - Priority and due dates
   - Task dependencies
   - Daily work hours
   - Hours-per-week limits
   - Existing calendar events
5. **Create Events**: Adds scheduled tasks as events to your Google Calendar

### Scheduling Algorithm

Tasks are scheduled based on:
1. Priority (higher priority tasks first)
2. Due date (tasks with earlier due dates first)
3. Dependencies (tasks are scheduled after their dependencies)
4. Available time slots (based on your work hours and existing events)
5. Buffer time between tasks
6. Hours-per-week limits

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
