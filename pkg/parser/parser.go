package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseProject parses a YAML file and returns a Project struct
func ParseProject(filePath string) (*Project, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var project Project
	if err := yaml.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &project, nil
}

// Validate checks if the project has valid data
func (p *Project) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if p.StartDate.IsZero() {
		return fmt.Errorf("start date is required")
	}
	if p.EndDate.IsZero() {
		return fmt.Errorf("end date is required")
	}
	if p.EndDate.Before(p.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	for i, task := range p.Tasks {
		if task.Name == "" {
			return fmt.Errorf("task %d: name is required", i)
		}
		if task.Duration == 0 {
			return fmt.Errorf("task %d: duration is required", i)
		}
	}

	return nil
}

// ParseSettings parses the settings from the project
func (p *Project) ParseSettings() error {
	// Parse work hours
	if p.Settings.WorkHours != "" {
		// WorkHours format: "09:00-17:00"
		// This will be parsed in the scheduler
	}

	// Parse buffer time
	if p.Settings.BufferTime != "" {
		duration, err := ParseDuration(p.Settings.BufferTime)
		if err != nil {
			return fmt.Errorf("invalid buffer time: %w", err)
		}
		// Store parsed duration - we'll need to add a field for this
		_ = duration
	}

	return nil
}
