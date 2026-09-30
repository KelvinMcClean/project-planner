package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/KelvinMcClean/project-planner/pkg/calendar"
	"github.com/KelvinMcClean/project-planner/pkg/parser"
	"github.com/KelvinMcClean/project-planner/pkg/scheduler"
)

var (
	projectFile string
	dryRun      bool
	verbose     bool
)

func init() {
	flag.StringVar(&projectFile, "file", "project.yml", "Path to the YAML project file")
	flag.BoolVar(&dryRun, "dry-run", false, "Preview the schedule without adding to calendar")
	flag.BoolVar(&verbose, "verbose", false, "Show detailed output")
	flag.Parse()
}

func main() {
	ctx := context.Background()

	// Parse the project file
	project, err := parser.ParseProject(projectFile)
	if err != nil {
		log.Fatalf("Failed to parse project file: %v", err)
	}

	if verbose {
		fmt.Printf("Project: %s\n", project.Name)
		fmt.Printf("Period: %s to %s\n", project.StartDate.Format("2006-01-02"), project.EndDate.Format("2006-01-02"))
		fmt.Printf("Tasks: %d\n", len(project.Tasks))
	}

	// Validate the project
	if err := project.Validate(); err != nil {
		log.Fatalf("Invalid project: %v", err)
	}

	// Create scheduler
	sched := scheduler.NewScheduler(project.StartDate, project.EndDate)

	// Configure scheduler from project settings
	if err := sched.Configure(project.Settings); err != nil {
		log.Printf("Warning: failed to configure scheduler: %v", err)
	}

	// If not in dry-run mode, get existing events from calendar
	if !dryRun {
		calClient, err := calendar.NewClient(ctx)
		if err != nil {
			log.Printf("Warning: failed to create calendar client: %v", err)
			log.Println("Continuing with scheduling (existing events will be ignored)")
		} else {
			// Get existing events to avoid
			existingSlots, err := calClient.GetExistingEvents(project.StartDate, project.EndDate)
			if err != nil {
				log.Printf("Warning: failed to get existing events: %v", err)
			} else {
				for _, slot := range existingSlots {
					sched.AddExistingEvent(slot.Start, slot.End)
				}
				if verbose {
					fmt.Printf("Found %d existing events to avoid\n", len(existingSlots))
				}
			}
		}
	}

	// Schedule tasks
	result := sched.Schedule(project.Tasks)

	// Print results
	fmt.Println("\n=== Schedule Results ===")
	fmt.Printf("Scheduled: %d tasks\n", len(result.ScheduledTasks))
	fmt.Printf("Unscheduled: %d tasks\n", len(result.Unscheduled))

	if len(result.Unscheduled) > 0 {
		fmt.Println("\nUnscheduled tasks:")
		for _, name := range result.Unscheduled {
			fmt.Printf("  - %s\n", name)
		}
	}

	if verbose || dryRun {
		fmt.Println("\nScheduled tasks:")
		for _, task := range result.ScheduledTasks {
			fmt.Printf("  %s: %s to %s (%v)\n",
				task.Name,
				task.Start.Format("2006-01-02 15:04"),
				task.End.Format("15:04"),
				task.Duration)
		}
	}

	// Add to calendar if not dry-run
	if !dryRun {
		calClient, err := calendar.NewClient(ctx)
		if err != nil {
			log.Fatalf("Failed to create calendar client: %v", err)
		}

		if len(result.ScheduledTasks) > 0 {
			fmt.Println("\nAdding scheduled tasks to Google Calendar...")
			_, err = calClient.CreateEvents(result.ScheduledTasks)
			if err != nil {
				log.Fatalf("Failed to create events: %v", err)
			}
			fmt.Printf("Successfully added %d events to calendar!\n", len(result.ScheduledTasks))
		}
	}
}
