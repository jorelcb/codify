package dto

// TemplateInfo representa información sobre un template
type TemplateInfo struct {
	ID          string
	Name        string
	Path        string
	Description string
	Language    string
	Tags        []string
	Variables   []string
}

// TemplateListResult representa el resultado de listar templates
type TemplateListResult struct {
	Templates []TemplateInfo
	Total     int
}