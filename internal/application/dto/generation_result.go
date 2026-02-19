package dto

// GenerationResult represents the result of an LLM-based context generation
type GenerationResult struct {
	OutputPath     string
	GeneratedFiles []string
	Model          string
	TokensIn       int
	TokensOut      int
}
