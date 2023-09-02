package summary

import (
	"context"
	"errors"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"github.com/sashabaranov/go-openai"
)

type OpenAI struct {
	c         *openai.Client
	gptConfig *config.GPT
}

func NewOpenAI(
	c *openai.Client,
	gptConfig *config.GPT,
) *OpenAI {
	return &OpenAI{c: c, gptConfig: gptConfig}
}

func (o *OpenAI) GetMsgSummary(ctx context.Context, msg entity.Message) (string, error) {
	resp, err := o.c.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{{
				Role:    openai.ChatMessageRoleUser,
				Content: o.gptConfig.FeedPrompt(msg.Body()),
			}},
			MaxTokens: o.gptConfig.MaxTokens,
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no responses returned")
	}

	return resp.Choices[0].Message.Content, nil
}
