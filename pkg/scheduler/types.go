package scheduler

import (
	"fmt"
	"strings"
	"time"
)

// TimeSlot represents a available time slot for scheduling
type TimeSlot struct {
	Start time.Time
	End   time.Time
}

// ScheduledTask represents a task that has been scheduled
type ScheduledTask struct {
	Name        string
	Description string
	Start       time.Time
	End         time.Time
	Duration    time.Duration
	Priority    int
	Tags        []string
	DependsOn   []string // names of tasks this task depends on
}

// ScheduleResult contains the scheduling results
type ScheduleResult struct {
	ScheduledTasks []ScheduledTask
	Unscheduled    []string // names of tasks that couldn't be scheduled
	Conflicts      []string // any conflicts detected
	DependencyIssues []string // tasks that couldn't be scheduled due to dependency issues
}

// WorkHours represents daily working hours for a specific day
type WorkHours struct {
	StartHour int
	StartMin  int
	EndHour   int
	EndMin    int
}

// DayWorkHours maps day names to their work hours
type DayWorkHours map[string]WorkHours

// ParseWorkHours parses a work hours string like "09:00-17:00"
func ParseWorkHours(wh string) (WorkHours, error) {
	var whs WorkHours
	_, err := fmt.Sscanf(wh, "%02d:%02d-%02d:%02d", &whs.StartHour, &whs.StartMin, &whs.EndHour, &whs.EndMin)
	if err != nil {
		return WorkHours{}, fmt.Errorf("invalid work hours format, use HH:MM-HH:MM: %w", err)
	}
	return whs, nil
}

// ParseDayWorkHoursMap parses a map of day names to work hours strings
func ParseDayWorkHoursMap(workHoursMap map[string]string) (DayWorkHours, error) {
	dwh := make(DayWorkHours)
	
	for day, hoursStr := range workHoursMap {
		wh, err := ParseWorkHours(hoursStr)
		if err != nil {
			return nil, fmt.Errorf("invalid work hours for %s: %w", day, err)
		}
		
		dwh[day] = wh
	}
	
	return dwh, nil
}

// TimeOfDay represents a time within a day (hour and minute)
type TimeOfDay struct {
	Hour int
	Min  int
}

// ToTime converts TimeOfDay to time.Time on a specific date
func (t TimeOfDay) ToTime(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour, t.Min, 0, 0, date.Location())
}

// FromTime extracts TimeOfDay from time.Time
func FromTime(t time.Time) TimeOfDay {
	return TimeOfDay{Hour: t.Hour(), Min: t.Minute()}
}
