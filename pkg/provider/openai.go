package provider

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/ihoegen/terra-inform/pkg/checks"
)

type OpenAIProvider struct {
	config Config
}

func NewOpenAIProvider(config Config) *OpenAIProvider {
	return &OpenAIProvider{
		config: config,
	}
}

func (p *OpenAIProvider) ProcessChecks(checks []checks.Check, input string) []CheckResult {
	return RunChecksInParallel(checks, input, p.processCheck)
}

func (p *OpenAIProvider) processCheck(check checks.Check, input string) (string, error) {
	client := openai.NewClient(
		option.WithAPIKey(p.config.APIKey),
	)

	chatCompletion, err := client.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(check.GetPrompt(input)),
			}),
			Model:     openai.F(p.config.ModelName),
			MaxTokens: openai.F(int64(500)),
		},
	)

	if err != nil {
		return "", fmt.Errorf("error processing check %s: %v", check.GetName(), err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
} 