package scheduler

import (
	"fmt"
	"sort"
	"time"

	"github.com/KelvinMcClean/project-planner/pkg/parser"
)

// Scheduler handles the scheduling logic
type Scheduler struct {
	WorkHours       WorkHours
	WorkDays        map[time.Weekday]bool
	BufferTime      time.Duration
	MaxTasksPerDay  int
	PreferredSlots  []TimeOfDay
	ExistingEvents  []TimeSlot // Time slots that are already busy
	ProjectStart    time.Time
	ProjectEnd      time.Time
}

// NewScheduler creates a new scheduler with default settings
func NewScheduler(start, end time.Time) *Scheduler {
	return &Scheduler{
		WorkHours:      WorkHours{StartHour: 9, StartMin: 0, EndHour: 17, EndMin: 0},
		WorkDays:       map[time.Weekday]bool{time.Monday: true, time.Tuesday: true, time.Wednesday: true, time.Thursday: true, time.Friday: true},
		BufferTime:      0,
		MaxTasksPerDay:  10,
		PreferredSlots:  nil,
		ExistingEvents:  nil,
		ProjectStart:    start,
		ProjectEnd:      end,
	}
}

// Configure sets up the scheduler from project settings
func (s *Scheduler) Configure(settings parser.Settings) error {
	if settings.WorkHours != "" {
		wh, err := ParseWorkHours(settings.WorkHours)
		if err != nil {
			return fmt.Errorf("failed to parse work hours: %w", err)
		}
		s.WorkHours = wh
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

	start := s.WorkHours.ToTime(day)
	end := time.Date(day.Year(), day.Month(), day.Day(), s.WorkHours.EndHour, s.WorkHours.EndMin, 0, 0, day.Location())

	// Subtract existing events
	availableSlots := []TimeSlot{{Start: start, End: end}}
	
	for _, event := range s.ExistingEvents {
		if event.Start.Year() == day.Year() && event.Start.Month() == day.Month() && event.Start.Day() == day.Day() {
			availableSlots = s.subtractSlot(availableSlots, event)
		}
	}

	return availableSlots
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
	}

	// Sort tasks by priority (higher first) and then by due date
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Priority != tasks[j].Priority {
			return tasks[i].Priority > tasks[j].Priority
		}
		if tasks[i].DueDate != nil && tasks[j].DueDate != nil {
			return tasks[i].DueDate.Before(*tasks[j].DueDate)
		}
		if tasks[i].DueDate != nil {
			return true
		}
		if tasks[j].DueDate != nil {
			return false
		}
		return i < j
	})

	// Get all available slots
	allSlots := s.GetAllAvailableSlots()

	// Schedule tasks
	for _, task := range tasks {
		scheduled := s.scheduleTask(task, allSlots)
		if scheduled.Start != (time.Time{}) {
			result.ScheduledTasks = append(result.ScheduledTasks, scheduled)
			// Remove the used slot
			allSlots = s.removeSlot(allSlots, TimeSlot{Start: scheduled.Start, End: scheduled.End})
		} else {
			result.Unscheduled = append(result.Unscheduled, task.Name)
		}
	}

	return result
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
