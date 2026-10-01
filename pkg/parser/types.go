package parser

import (
	"fmt"
	"time"
)

// Task represents a single task from the YAML file
type Task struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description,omitempty"`
	Duration    time.Duration `yaml:"duration"`
	Priority    int           `yaml:"priority,omitempty"`
	DueDate     *time.Time    `yaml:"due_date,omitempty"`
	Tags        []string      `yaml:"tags,omitempty"`
	DependsOn   []string      `yaml:"depends_on,omitempty"` // names of tasks this task depends on
}

// Project represents the entire project with tasks and settings
type Project struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description,omitempty"`
	StartDate   time.Time `yaml:"start_date"`
	EndDate     time.Time `yaml:"end_date"`
	Tasks       []Task    `yaml:"tasks"`
	Settings    Settings  `yaml:"settings,omitempty"`
}

// Settings contains scheduling preferences
type Settings struct {
	WorkHours       map[string]string `yaml:"work_hours,omitempty"`       // e.g., {"mon": "09:00-17:00", "tue": "10:00-18:00"}
	WorkDays        []string         `yaml:"work_days,omitempty"`        // e.g., ["mon", "tue", "wed", "thu", "fri"]
	BufferTime      string           `yaml:"buffer_time,omitempty"`      // e.g., "30m"
	MaxTasksPerDay  int              `yaml:"max_tasks_per_day,omitempty"` // max tasks to schedule per day
	HoursPerWeek    float64          `yaml:"hours_per_week,omitempty"`   // maximum hours to schedule per week
	PreferredSlots  []string         `yaml:"preferred_slots,omitempty"`  // preferred time slots
	AvoidOverlap    bool             `yaml:"avoid_overlap,omitempty"`    // avoid overlapping with existing events
}

// ParseDuration parses a duration string like "2h", "30m", "1h30m" into time.Duration
func ParseDuration(s string) (time.Duration, error) {
	var total time.Duration
	var numStr string
	
	for _, c := range s {
		if c >= '0' && c <= '9' {
			numStr += string(c)
		} else {
			if numStr != "" {
				num := 0
				_, err := fmt.Sscanf(numStr, "%d", &num)
				if err != nil {
					return 0, fmt.Errorf("invalid number in duration: %s", numStr)
				}
				
				switch c {
				case 'h':
					total += time.Duration(num) * time.Hour
				case 'm':
					total += time.Duration(num) * time.Minute
				case 's':
					total += time.Duration(num) * time.Second
				case 'd':
					total += time.Duration(num) * 24 * time.Hour
				default:
					return 0, fmt.Errorf("unknown duration unit: %c", c)
				}
				numStr = ""
			}
		}
	}
	
	// Handle case where duration ends with a number
	if numStr != "" {
		num := 0
		_, err := fmt.Sscanf(numStr, "%d", &num)
		if err != nil {
			return 0, fmt.Errorf("invalid number in duration: %s", numStr)
		}
		total += time.Duration(num) * time.Minute
	}
	
	return total, nil
}
