package scheduler

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/KelvinMcClean/project-planner/pkg/parser"
)

// Scheduler handles the scheduling logic
type Scheduler struct {
	DayWorkHours    DayWorkHours
	WorkDays        map[time.Weekday]bool
	BufferTime      time.Duration
	MaxTasksPerDay  int
	HoursPerWeek    float64
	PreferredSlots  []TimeOfDay
	ExistingEvents  []TimeSlot // Time slots that are already busy
	ProjectStart    time.Time
	ProjectEnd      time.Time
	ScheduledTasks   map[string]ScheduledTask // track scheduled tasks by name for dependency resolution
}

// NewScheduler creates a new scheduler with default settings
func NewScheduler(start, end time.Time) *Scheduler {
	// Default work hours for all days
	defaultHours := WorkHours{StartHour: 9, StartMin: 0, EndHour: 17, EndMin: 0}
	defaultDayHours := make(DayWorkHours)
	defaultDayHours["mon"] = defaultHours
	defaultDayHours["tue"] = defaultHours
	defaultDayHours["wed"] = defaultHours
	defaultDayHours["thu"] = defaultHours
	defaultDayHours["fri"] = defaultHours
	defaultDayHours["sat"] = defaultHours
	defaultDayHours["sun"] = defaultHours

	return &Scheduler{
		DayWorkHours:    defaultDayHours,
		WorkDays:        map[time.Weekday]bool{time.Monday: true, time.Tuesday: true, time.Wednesday: true, time.Thursday: true, time.Friday: true},
		BufferTime:      0,
		MaxTasksPerDay:  10,
		HoursPerWeek:    0, // 0 means no limit
		PreferredSlots:  nil,
		ExistingEvents:  nil,
		ProjectStart:    start,
		ProjectEnd:      end,
		ScheduledTasks:  make(map[string]ScheduledTask),
	}
}

// Configure sets up the scheduler from project settings
func (s *Scheduler) Configure(settings parser.Settings) error {
	// Handle daily work hours
	if len(settings.WorkHours) > 0 {
		// If WorkHours is a map (daily customizable)
		if workHoursMap, ok := settings.WorkHours.(map[string]string); ok {
			dwh, err := ParseDayWorkHoursMap(workHoursMap)
			if err != nil {
				return fmt.Errorf("failed to parse daily work hours: %w", err)
			}
			s.DayWorkHours = dwh
		} else if whStr, ok := settings.WorkHours.(string); ok && whStr != "" {
			// Single work hours string - apply to all days
			wh, err := ParseWorkHours(whStr)
			if err != nil {
				return fmt.Errorf("failed to parse work hours: %w", err)
			}
			for day := range s.DayWorkHours {
				s.DayWorkHours[day] = wh
			}
		}
	}

	if len(settings.WorkDays) > 0 {
		s.WorkDays = make(map[time.Weekday]bool)
		for _, day := range settings.WorkDays {
			switch day {
			case "mon":
				s.WorkDays[time.Monday] = true
			case "tue":
				s.WorkDays[time.Tuesday] = true
			case "wed":
				s.WorkDays[time.Wednesday] = true
			case "thu":
				s.WorkDays[time.Thursday] = true
			case "fri":
				s.WorkDays[time.Friday] = true
			case "sat":
				s.WorkDays[time.Saturday] = true
			case "sun":
				s.WorkDays[time.Sunday] = true
			}
		}
	}

	if settings.BufferTime != "" {
		duration, err := parser.ParseDuration(settings.BufferTime)
		if err != nil {
			return fmt.Errorf("failed to parse buffer time: %w", err)
		}
		s.BufferTime = duration
	}

	if settings.MaxTasksPerDay > 0 {
		s.MaxTasksPerDay = settings.MaxTasksPerDay
	}

	if settings.HoursPerWeek > 0 {
		s.HoursPerWeek = settings.HoursPerWeek
	}

	return nil
}

// GetWorkDaySlots returns available time slots for a given day
func (s *Scheduler) GetWorkDaySlots(day time.Time) []TimeSlot {
	var slots []TimeSlot

	// Check if this day is a work day
	if !s.WorkDays[day.Weekday()] {
		return slots
	}

	// Check if day is within project range
	if day.Before(s.ProjectStart) || day.After(s.ProjectEnd) {
		return slots
	}

	// Get work hours for this specific day
	wh, hasCustom := s.getWorkHoursForDay(day)
	if !hasCustom {
		// Fall back to default
		wh = WorkHours{StartHour: 9, StartMin: 0, EndHour: 17, EndMin: 0}
	}

	start := wh.ToTime(day)
	end := time.Date(day.Year(), day.Month(), day.Day(), wh.EndHour, wh.EndMin, 0, 0, day.Location())

	// Subtract existing events
	availableSlots := []TimeSlot{{Start: start, End: end}}
	
	for _, event := range s.ExistingEvents {
		if event.Start.Year() == day.Year() && event.Start.Month() == day.Month() && event.Start.Day() == day.Day() {
			availableSlots = s.subtractSlot(availableSlots, event)
		}
	}

	return availableSlots
}

// getWorkHoursForDay returns the work hours for a specific day
func (s *Scheduler) getWorkHoursForDay(day time.Time) (WorkHours, bool) {
	weekday := day.Weekday()
	
	// Try to find exact day name match
	dayNames := []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}
	for _, name := range dayNames {
		if wh, exists := s.DayWorkHours[name]; exists {
			// Map day name to weekday
			var targetDay time.Weekday
			switch name {
			case "mon": targetDay = time.Monday
			case "tue": targetDay = time.Tuesday
			case "wed": targetDay = time.Wednesday
			case "thu": targetDay = time.Thursday
			case "fri": targetDay = time.Friday
			case "sat": targetDay = time.Saturday
			case "sun": targetDay = time.Sunday
			}
			if targetDay == weekday {
				return wh, true
			}
		}
	}
	
	return WorkHours{}, false
}

// subtractSlot removes an occupied slot from available slots
func (s *Scheduler) subtractSlot(available []TimeSlot, occupied TimeSlot) []TimeSlot {
	var result []TimeSlot
	
	for _, slot := range available {
		// No overlap
		if slot.End <= occupied.Start || slot.Start >= occupied.End {
			result = append(result, slot)
			continue
		}

		// Occupied slot is completely within this available slot
		if occupied.Start >= slot.Start && occupied.End <= slot.End {
			if occupied.Start > slot.Start {
				result = append(result, TimeSlot{Start: slot.Start, End: occupied.Start})
			}
			if occupied.End < slot.End {
				result = append(result, TimeSlot{Start: occupied.End, End: slot.End})
			}
			continue
		}

		// Occupied slot overlaps with the start of available slot
		if occupied.Start <= slot.Start && occupied.End > slot.Start {
			result = append(result, TimeSlot{Start: occupied.End, End: slot.End})
			continue
		}

		// Occupied slot overlaps with the end of available slot
		if occupied.Start < slot.End && occupied.End >= slot.End {
			result = append(result, TimeSlot{Start: slot.Start, End: occupied.Start})
			continue
		}
	}

	return result
}

// ToTime converts WorkHours to time.Time on a specific date
func (wh WorkHours) ToTime(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), wh.StartHour, wh.StartMin, 0, 0, date.Location())
}

// Schedule schedules tasks into available time slots
func (s *Scheduler) Schedule(tasks []parser.Task) ScheduleResult {
	result := ScheduleResult{
		ScheduledTasks: []ScheduledTask{},
		Unscheduled:    []string{},
		Conflicts:      []string{},
		DependencyIssues: []string{},
	}

	// Build task name to task map for dependency resolution
	taskMap := make(map[string]parser.Task)
	for _, task := range tasks {
		taskMap[task.Name] = task
	}

	// Sort tasks by priority (higher first) and then by due date
	sortedTasks := make([]parser.Task, len(tasks))
	copy(sortedTasks, tasks)
	
	sort.Slice(sortedTasks, func(i, j int) bool {
		if sortedTasks[i].Priority != sortedTasks[j].Priority {
			return sortedTasks[i].Priority > sortedTasks[j].Priority
		}
		if sortedTasks[i].DueDate != nil && sortedTasks[j].DueDate != nil {
			return sortedTasks[i].DueDate.Before(*sortedTasks[j].DueDate)
		}
		if sortedTasks[i].DueDate != nil {
			return true
		}
		if sortedTasks[j].DueDate != nil {
			return false
		}
		return i < j
	})

	// Track scheduled tasks
	scheduledMap := make(map[string]ScheduledTask)
	
	// Get all available slots
	allSlots := s.GetAllAvailableSlots()
	
	// Track hours scheduled per week for limits
	weeklyHours := make(map[int]time.Duration) // year-week -> hours

	// Schedule tasks - need to handle dependencies
	for _, task := range sortedTasks {
		// Check if all dependencies are scheduled
		allDepsMet := true
		for _, depName := range task.DependsOn {
			if _, exists := scheduledMap[depName]; !exists {
				allDepsMet = false
				break
			}
		}
		
		if !allDepsMet {
			// Try to schedule dependencies first
			depsScheduled := s.tryScheduleDependencies(task.DependsOn, taskMap, scheduledMap, &allSlots, &result, weeklyHours)
			if !depsScheduled {
				result.DependencyIssues = append(result.DependencyIssues, task.Name)
				continue
			}
			// Dependencies are now scheduled, try again
			allDepsMet = true
			for _, depName := range task.DependsOn {
				if _, exists := scheduledMap[depName]; !exists {
					allDepsMet = false
					break
				}
			}
		}

		if !allDepsMet {
			result.DependencyIssues = append(result.DependencyIssues, task.Name)
			continue
		}

		// Check hours per week limit
		year, week := day.ISOWeek()
		if s.HoursPerWeek > 0 {
			currentHours := weeklyHours[year*100+week]
			if float64(currentHours.Hours())+task.Duration.Hours() > s.HoursPerWeek {
				result.Unscheduled = append(result.Unscheduled, task.Name)
				continue
			}
		}

		scheduled := s.scheduleTask(task, allSlots)
		if scheduled.Start != (time.Time{}) {
			// Check if this would exceed hours per week
			if s.HoursPerWeek > 0 {
				year, week := scheduled.Start.ISOWeek()
				currentHours := weeklyHours[year*100+week]
				if float64(currentHours.Hours())+task.Duration.Hours() > s.HoursPerWeek {
					result.Unscheduled = append(result.Unscheduled, task.Name)
					continue
				}
				weeklyHours[year*100+week] = currentHours + task.Duration
			}
			
			result.ScheduledTasks = append(result.ScheduledTasks, scheduled)
			scheduledMap[task.Name] = scheduled
			s.ScheduledTasks[task.Name] = scheduled
			// Remove the used slot
			allSlots = s.removeSlot(allSlots, TimeSlot{Start: scheduled.Start, End: scheduled.End})
		} else {
			result.Unscheduled = append(result.Unscheduled, task.Name)
		}
	}

	return result
}

// tryScheduleDependencies attempts to schedule all dependencies for a task
func (s *Scheduler) tryScheduleDependencies(depNames []string, taskMap map[string]parser.Task, scheduledMap map[string]ScheduledTask, allSlots *[]TimeSlot, result *ScheduleResult, weeklyHours map[int]time.Duration) bool {
	allScheduled := true
	
	for _, depName := range depNames {
		if _, exists := scheduledMap[depName]; exists {
			continue
		}
		
		depTask, exists := taskMap[depName]
		if !exists {
			result.Conflicts = append(result.Conflicts, fmt.Sprintf("dependency '%s' not found", depName))
			allScheduled = false
			continue
		}
		
		// Schedule the dependency
		scheduled := s.scheduleTask(depTask, *allSlots)
		if scheduled.Start != (time.Time{}) {
			// Check hours per week limit
			if s.HoursPerWeek > 0 {
				year, week := scheduled.Start.ISOWeek()
				currentHours := weeklyHours[year*100+week]
				if float64(currentHours.Hours())+depTask.Duration.Hours() > s.HoursPerWeek {
					result.Unscheduled = append(result.Unscheduled, depTask.Name)
					allScheduled = false
					continue
				}
				weeklyHours[year*100+week] = currentHours + depTask.Duration
			}
			
			result.ScheduledTasks = append(result.ScheduledTasks, scheduled)
			scheduledMap[depTask.Name] = scheduled
			s.ScheduledTasks[depTask.Name] = scheduled
			*allSlots = s.removeSlot(*allSlots, TimeSlot{Start: scheduled.Start, End: scheduled.End})
		} else {
			result.Unscheduled = append(result.Unscheduled, depTask.Name)
			allScheduled = false
		}
	}
	
	return allScheduled
}

// GetAllAvailableSlots returns all available time slots in the project period
func (s *Scheduler) GetAllAvailableSlots() []TimeSlot {
	var allSlots []TimeSlot

	current := s.ProjectStart
	for current.Before(s.ProjectEnd) || current.Equal(s.ProjectEnd) {
		daySlots := s.GetWorkDaySlots(current)
		allSlots = append(allSlots, daySlots...)
		current = current.Add(24 * time.Hour)
	}

	return allSlots
}

// scheduleTask tries to schedule a single task
func (s *Scheduler) scheduleTask(task parser.Task, slots []TimeSlot) ScheduledTask {
	// Sort slots by start time
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].Start.Before(slots[j].Start)
	})

	// Try to find a slot that fits the task
	for _, slot := range slots {
		if slot.End.Sub(slot.Start) >= task.Duration+s.BufferTime {
			start := slot.Start
			if s.BufferTime > 0 {
				start = start.Add(s.BufferTime / 2)
			}
			end := start.Add(task.Duration)
			
			// Check if end is within the slot
			if end.After(slot.End) {
				continue
			}

			return ScheduledTask{
				Name:        task.Name,
				Description: task.Description,
				Start:       start,
				End:         end,
				Duration:    task.Duration,
				Priority:    task.Priority,
				Tags:        task.Tags,
				DependsOn:   task.DependsOn,
			}
		}
	}

	return ScheduledTask{}
}

// removeSlot removes a specific time slot from the list
func (s *Scheduler) removeSlot(slots []TimeSlot, toRemove TimeSlot) []TimeSlot {
	var result []TimeSlot
	for _, slot := range slots {
		if slot.Start != toRemove.Start || slot.End != toRemove.End {
			result = append(result, slot)
		}
	}
	return result
}

// AddExistingEvent adds an existing calendar event to avoid
func (s *Scheduler) AddExistingEvent(start, end time.Time) {
	s.ExistingEvents = append(s.ExistingEvents, TimeSlot{Start: start, End: end})
}

// day.ISOWeek is not available, so we implement our own
func (d time.Time) ISOWeek() (year, week int) {
	year, week = d.Year(), 0
	// Simple implementation - week starts on Monday
	// This is a simplified version
	dayOfYear := d.YearDay()
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, d.Location())
	jan1Weekday := int(jan1.Weekday())
	// Adjust for Monday start
	week = (dayOfYear + jan1Weekday + 6) / 7
	if jan1Weekday == time.Sunday {
		week = (dayOfYear + 6) / 7
	}
	return
}
