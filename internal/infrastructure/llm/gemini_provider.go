package llm

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/genai"

	"github.com/jorelcb/codify/internal/domain/service"
)

const defaultGeminiModel = "gemini-3.1-pro-preview"

// GeminiProvider implements service.LLMProvider using the Google Gemini API.
type GeminiProvider struct {
	client        *genai.Client
	model         string
	promptBuilder *PromptBuilder
	progressOut   io.Writer
}

// NewGeminiProvider creates a new GeminiProvider.
// If apiKey is empty, the SDK will use the GEMINI_API_KEY or GOOGLE_API_KEY env var.
// If model is empty, defaults to gemini-3.1-pro-preview.
// If progressOut is non-nil, progress messages will be written to it.
func NewGeminiProvider(ctx context.Context, apiKey string, model string, progressOut io.Writer) (*GeminiProvider, error) {
	config := &genai.ClientConfig{
		Backend: genai.BackendGeminiAPI,
	}
	if apiKey != "" {
		config.APIKey = apiKey
	}

	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	if model == "" {
		model = defaultGeminiModel
	}

	return &GeminiProvider{
		client:        client,
		model:         model,
		promptBuilder: NewPromptBuilder(),
		progressOut:   progressOut,
	}, nil
}

// GenerateContext generates all context files by making one API call per file.
func (p *GeminiProvider) GenerateContext(ctx context.Context, req service.GenerationRequest) (*service.GenerationResponse, error) {
	var files []service.GeneratedFile
	var totalIn, totalOut int

	for i, guide := range req.TemplateGuides {
		outputName := FileOutputName(guide.Name)

		if p.progressOut != nil {
			fmt.Fprintf(p.progressOut, "  [%d/%d] Generating %s...", i+1, len(req.TemplateGuides), outputName)
		}

		content, tokensIn, tokensOut, err := p.generateSingleFile(ctx, req, guide)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s: %w", outputName, err)
		}

		totalIn += tokensIn
		totalOut += tokensOut

		if p.progressOut != nil {
			fmt.Fprintf(p.progressOut, " done (%d tokens)\n", tokensOut)
		}

		files = append(files, service.GeneratedFile{
			Name:    outputName,
			Content: content,
		})
	}

	return &service.GenerationResponse{
		Files:     files,
		Model:     p.model,
		TokensIn:  totalIn,
		TokensOut: totalOut,
	}, nil
}

// generateSingleFile makes one streaming API call to generate a single context file.
func (p *GeminiProvider) generateSingleFile(
	ctx context.Context,
	req service.GenerationRequest,
	guide service.TemplateGuide,
) (content string, tokensIn int, tokensOut int, err error) {
	var systemPrompt string
	var userMessage string
	switch req.Mode {
	case "spec":
		systemPrompt = p.promptBuilder.BuildSpecSystemPrompt(req.ExistingContext, req.Locale)
		userMessage = p.promptBuilder.BuildUserMessageForFile(req, guide)
	case "skills":
		systemPrompt = p.promptBuilder.BuildPersonalizedSkillsSystemPrompt(guide.Name, req.Target, req.Locale, req.ProjectContext)
		userMessage = p.promptBuilder.BuildSkillsUserMessage(guide, req.Target)
	case "workflows":
		if req.Target == "claude" {
			systemPrompt = p.promptBuilder.BuildClaudeWorkflowSystemPrompt(guide.Name, req.Locale, req.ProjectContext)
		} else {
			systemPrompt = p.promptBuilder.BuildPersonalizedWorkflowsSystemPrompt(guide.Name, req.Locale, req.ProjectContext)
		}
		userMessage = p.promptBuilder.BuildWorkflowsUserMessage(guide, req.Target)
	default:
		systemPrompt = p.promptBuilder.BuildSystemPromptForFile(guide.Name, req.Locale)
		userMessage = p.promptBuilder.BuildUserMessageForFile(req, guide)
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
		MaxOutputTokens:   16000,
	}

	var textBuilder strings.Builder
	var inTokens, outTokens int32

	for resp, err := range p.client.Models.GenerateContentStream(
		ctx,
		p.model,
		genai.Text(userMessage),
		config,
	) {
		if err != nil {
			return "", 0, 0, fmt.Errorf("streaming failed: %w", err)
		}

		if resp.UsageMetadata != nil {
			inTokens = resp.UsageMetadata.PromptTokenCount
			outTokens = resp.UsageMetadata.CandidatesTokenCount
		}

		for _, candidate := range resp.Candidates {
			if candidate.Content != nil {
				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						textBuilder.WriteString(part.Text)
					}
				}
			}
		}
	}

	text := textBuilder.String()
	if text == "" {
		return "", 0, 0, fmt.Errorf("empty response from LLM")
	}

	return text, int(inTokens), int(outTokens), nil
}
