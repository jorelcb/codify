package dto

import "time"

// ProjectInfo represents information about a generated project
type ProjectInfo struct {
	ID           string
	Name         string
	Language     string
	Type         string
	Architecture string
	OutputPath   string
	Capabilities []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ProjectListResult represents the result of listing projects
type ProjectListResult struct {
	Projects []ProjectInfo
	Total    int
}
