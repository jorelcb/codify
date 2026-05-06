package llm

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/jorelcb/codify/internal/domain/service"
)

const defaultModel = "claude-sonnet-4-6"

// AnthropicProvider implements service.LLMProvider using the Anthropic Claude API.
type AnthropicProvider struct {
	client        anthropic.Client
	model         string
	promptBuilder *PromptBuilder
	progressOut   io.Writer
}

// NewAnthropicProvider creates a new AnthropicProvider.
// If apiKey is empty, the SDK will use the ANTHROPIC_API_KEY env var.
// If model is empty, defaults to claude-sonnet-4-6.
// If progressOut is non-nil, progress messages will be written to it.
func NewAnthropicProvider(apiKey string, model string, progressOut io.Writer) *AnthropicProvider {
	var opts []option.RequestOption
	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	}

	client := anthropic.NewClient(opts...)

	if model == "" {
		model = defaultModel
	}

	return &AnthropicProvider{
		client:        client,
		model:         model,
		promptBuilder: NewPromptBuilder(),
		progressOut:   progressOut,
	}
}

// GenerateContext generates all context files by making one API call per file.
// This avoids JSON truncation issues and provides per-file progress.
func (p *AnthropicProvider) GenerateContext(ctx context.Context, req service.GenerationRequest) (*service.GenerationResponse, error) {
	start := time.Now()
	var files []service.GeneratedFile
	var totalIn, totalOut int
	success := true

	for i, guide := range req.TemplateGuides {
		outputName := FileOutputName(guide.Name)

		if p.progressOut != nil {
			fmt.Fprintf(p.progressOut, "  [%d/%d] Generating %s...", i+1, len(req.TemplateGuides), outputName)
		}

		content, tokensIn, tokensOut, err := p.generateSingleFile(ctx, req, guide)
		totalIn += tokensIn
		totalOut += tokensOut

		if err != nil {
			success = false
			recordUsage("anthropic", p.model, commandFromMode(req.Mode), totalIn, totalOut, time.Since(start), false)
			return nil, fmt.Errorf("failed to generate %s: %w", outputName, err)
		}

		if p.progressOut != nil {
			fmt.Fprintf(p.progressOut, " done (%d tokens)\n", tokensOut)
		}

		validation := ValidateOutput(content, req.Mode, outputName)
		if validation.Fatal {
			recordUsage("anthropic", p.model, commandFromMode(req.Mode), totalIn, totalOut, time.Since(start), false)
			return nil, fmt.Errorf("output for %s was rejected by validator: %v", outputName, validation.Warnings)
		}
		emitValidationFeedback(p.progressOut, outputName, validation)

		files = append(files, service.GeneratedFile{
			Name:    outputName,
			Content: content,
		})
	}

	recordUsage("anthropic", p.model, commandFromMode(req.Mode), totalIn, totalOut, time.Since(start), success)

	return &service.GenerationResponse{
		Files:     files,
		Model:     p.model,
		TokensIn:  totalIn,
		TokensOut: totalOut,
	}, nil
}

// generateSingleFile makes one streaming API call to generate a single context file.
func (p *AnthropicProvider) generateSingleFile(
	ctx context.Context,
	req service.GenerationRequest,
	guide service.TemplateGuide,
) (content string, tokensIn int, tokensOut int, err error) {
	// Modes that personalize against a user-provided project context require it non-empty.
	// Without it the LLM has nothing to anchor on and tends to invent stack details.
	needsContext := req.Mode == "skills" || req.Mode == "workflows" || req.Mode == "workflow-skills"
	if needsContext && req.ProjectContext == "" {
		return "", 0, 0, fmt.Errorf("mode %q requires non-empty ProjectContext", req.Mode)
	}

	var systemPrompt string
	var userMessage string
	switch req.Mode {
	case "spec":
		systemPrompt = p.promptBuilder.BuildSpecSystemPrompt(req.ExistingContext, req.Locale)
		userMessage = p.promptBuilder.BuildUserMessageForFile(req, guide)
	case "skills":
		systemPrompt = p.promptBuilder.BuildPersonalizedSkillsSystemPrompt(guide.Name, req.Target, req.Locale, req.ProjectContext)
		userMessage = p.promptBuilder.BuildSkillsUserMessage(guide, req.Target)
	case "workflow-skills":
		systemPrompt = p.promptBuilder.BuildWorkflowSkillSystemPrompt(guide.Name, req.Locale, req.ProjectContext)
		userMessage = p.promptBuilder.BuildWorkflowSkillUserMessage(guide)
	case "workflows":
		systemPrompt = p.promptBuilder.BuildPersonalizedWorkflowsSystemPrompt(guide.Name, req.Locale, req.ProjectContext)
		userMessage = p.promptBuilder.BuildWorkflowsUserMessage(guide, req.Target)
	case "analyze":
		systemPrompt = p.promptBuilder.BuildAnalyzeSystemPromptForFile(guide.Name, req.Locale)
		userMessage = p.promptBuilder.BuildUserMessageForFile(req, guide)
	default:
		if req.Mode != "" && req.Mode != "generate" {
			return "", 0, 0, fmt.Errorf("unknown generation mode: %q", req.Mode)
		}
		systemPrompt = p.promptBuilder.BuildSystemPromptForFile(guide.Name, req.Locale)
		userMessage = p.promptBuilder.BuildUserMessageForFile(req, guide)
	}

	// Mark the system prompt as cacheable so subsequent calls within the
	// same generation (one call per template guide) reuse the prompt cache
	// instead of re-billing the full system prompt every time.
	stream := p.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 16000,
		System: []anthropic.TextBlockParam{
			{
				Text:         systemPrompt,
				CacheControl: anthropic.NewCacheControlEphemeralParam(),
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)),
		},
	})

	var textBuilder strings.Builder
	var inTokens, outTokens int64

	for stream.Next() {
		event := stream.Current()

		switch evt := event.AsAny().(type) {
		case anthropic.MessageStartEvent:
			inTokens = evt.Message.Usage.InputTokens
		case anthropic.ContentBlockDeltaEvent:
			switch delta := evt.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				textBuilder.WriteString(delta.Text)
			}
		case anthropic.MessageDeltaEvent:
			outTokens = evt.Usage.OutputTokens
		}
	}

	if err := stream.Err(); err != nil {
		return "", 0, 0, fmt.Errorf("streaming failed: %w", err)
	}

	text := textBuilder.String()
	if text == "" {
		return "", 0, 0, fmt.Errorf("empty response from LLM")
	}

	return text, int(inTokens), int(outTokens), nil
}

// EvaluatePrompt implements service.LLMProvider for one-shot prompt evaluation.
// Used by `audit --with-llm` and other lifecycle commands that need a single
// prompt→response cycle without the multi-file template flow.
//
// Token usage is recorded automatically via the same shim used by GenerateContext.
func (p *AnthropicProvider) EvaluatePrompt(ctx context.Context, req service.EvaluationRequest) (*service.EvaluationResponse, error) {
	start := time.Now()
	maxTokens := int64(req.MaxTokens)
	if maxTokens == 0 {
		maxTokens = 4000 // sensible default for audit-sized responses
	}

	systemBlock := anthropic.TextBlockParam{Text: req.SystemPrompt}
	if req.CacheableSystem {
		// Mark the system prompt as cacheable so a sequence of EvaluatePrompt
		// calls with the same SystemPrompt within the 5-minute TTL window
		// reuses the prompt cache instead of re-billing. Used by the marker
		// enricher (one call per generated file with markers) — high-traffic
		// case where the system prompt is identical across calls.
		systemBlock.CacheControl = anthropic.NewCacheControlEphemeralParam()
	}
	stream := p.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: maxTokens,
		System:    []anthropic.TextBlockParam{systemBlock},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.UserPrompt)),
		},
	})

	var textBuilder strings.Builder
	var inTokens, outTokens int64
	for stream.Next() {
		event := stream.Current()
		switch evt := event.AsAny().(type) {
		case anthropic.MessageStartEvent:
			inTokens = evt.Message.Usage.InputTokens
		case anthropic.ContentBlockDeltaEvent:
			if delta, ok := evt.Delta.AsAny().(anthropic.TextDelta); ok {
				textBuilder.WriteString(delta.Text)
			}
		case anthropic.MessageDeltaEvent:
			outTokens = evt.Usage.OutputTokens
		}
	}

	cmd := req.Command
	if cmd == "" {
		cmd = "evaluate"
	}

	if err := stream.Err(); err != nil {
		recordUsage("anthropic", p.model, cmd, int(inTokens), int(outTokens), time.Since(start), false)
		return nil, fmt.Errorf("streaming failed: %w", err)
	}

	text := textBuilder.String()
	recordUsage("anthropic", p.model, cmd, int(inTokens), int(outTokens), time.Since(start), text != "")
	if text == "" {
		return nil, fmt.Errorf("empty response from LLM")
	}

	return &service.EvaluationResponse{
		Text:      text,
		Model:     p.model,
		TokensIn:  int(inTokens),
		TokensOut: int(outTokens),
	}, nil
}
