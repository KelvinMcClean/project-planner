package calendar

import (
	"fmt"
	"time"

	"github.com/KelvinMcClean/project-planner/pkg/scheduler"
	"google.golang.org/api/calendar/v3"
)

// CalendarID is the ID of the calendar to use (primary by default)
const CalendarID = "primary"

// GetExistingEvents retrieves existing events from the calendar
func (c *Client) GetExistingEvents(start, end time.Time) ([]scheduler.TimeSlot, error) {
	var slots []scheduler.TimeSlot

	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	events, err := c.Service.Events.List(CalendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(startStr).
		TimeMax(endStr).
		MaxResults(2500).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	for _, item := range events.Items {
		startTime, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			// Try with date format
			startTime, err = time.Parse("2006-01-02", item.Start.Date)
			if err != nil {
				continue
			}
		}

		endTime, err := time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			// Try with date format
			endTime, err = time.Parse("2006-01-02", item.End.Date)
			if err != nil {
				continue
			}
		}

		slots = append(slots, scheduler.TimeSlot{
			Start: startTime,
			End:   endTime,
		})
	}

	return slots, nil
}

// CreateEvent creates a new calendar event from a scheduled task
func (c *Client) CreateEvent(task scheduler.ScheduledTask) (*calendar.Event, error) {
	startStr := task.Start.Format(time.RFC3339)
	endStr := task.End.Format(time.RFC3339)

	event := &calendar.Event{
		Summary:     task.Name,
		Description: task.Description,
		Start: &calendar.EventDateTime{
			DateTime: startStr,
			TimeZone: task.Start.Location().String(),
		},
		End: &calendar.EventDateTime{
			DateTime: endStr,
			TimeZone: task.End.Location().String(),
		},
	}

	// Add tags as extended properties if present
	if len(task.Tags) > 0 {
		event.ExtendedProperties = &calendar.EventExtendedProperties{
			Private: map[string]string{
				"tags": fmt.Sprintf("%v", task.Tags),
			},
		}
	}

	createdEvent, err := c.Service.Events.Insert(CalendarID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return createdEvent, nil
}

// CreateEvents creates multiple events from scheduled tasks
func (c *Client) CreateEvents(tasks []scheduler.ScheduledTask) ([]*calendar.Event, error) {
	var events []*calendar.Event

	for _, task := range tasks {
		event, err := c.CreateEvent(task)
		if err != nil {
			return events, fmt.Errorf("failed to create event for task %s: %w", task.Name, err)
		}
		events = append(events, event)
	}

	return events, nil
}

// DeleteEvent deletes an event by ID
func (c *Client) DeleteEvent(eventID string) error {
	err := c.Service.Events.Delete(CalendarID, eventID).Do()
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// GetFreeBusy returns free/busy information for a time range
func (c *Client) GetFreeBusy(start, end time.Time) ([]scheduler.TimeSlot, error) {
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	freeBusy, err := c.Service.FreeBusy.Query().
		TimeMin(startStr).
		TimeMax(endStr).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get free/busy: %w", err)
	}

	var busySlots []scheduler.TimeSlot
	for _, fb := range freeBusy.Calendars {
		for _, item := range fb.Busy {
			startTime, err := time.Parse(time.RFC3339, item.Start)
			if err != nil {
				continue
			}
			endTime, err := time.Parse(time.RFC3339, item.End)
			if err != nil {
				continue
			}
			busySlots = append(busySlots, scheduler.TimeSlot{
				Start: startTime,
				End:   endTime,
			})
		}
	}

	return busySlots, nil
}
