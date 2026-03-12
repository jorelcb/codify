package llm

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/jorelcb/ai-context-generator/internal/domain/service"
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
func (p *AnthropicProvider) generateSingleFile(
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
		systemPrompt = p.promptBuilder.BuildSkillsSystemPrompt(guide.Name, req.Target, req.Locale)
		userMessage = p.promptBuilder.BuildSkillsUserMessage(guide, req.Target)
	default:
		systemPrompt = p.promptBuilder.BuildSystemPromptForFile(guide.Name, req.Locale)
		userMessage = p.promptBuilder.BuildUserMessageForFile(req, guide)
	}

	stream := p.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 16000,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
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
