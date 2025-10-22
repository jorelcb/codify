package dto

import "time"

// ProjectInfo representa información sobre un proyecto generado
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

// ProjectListResult representa el resultado de listar proyectos
type ProjectListResult struct {
	Projects []ProjectInfo
	Total    int
}