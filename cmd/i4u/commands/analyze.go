package commands

import (
	"context"
	"github.com/fadyat/i4u/internal/config"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
)

func analyze(gptConfig *config.GPT) *cobra.Command {
	return &cobra.Command{
		Use:   "analyze",
		Short: "Analyze a text",
		Long: `
This command will analyze a text and return a summary of it.
Used for testing purposes, this command will be in the some
of the job's, that run after receiving an email, logic.
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mailMessage := args[0]

			client := openai.NewClient(gptConfig.OpenAIKey)
			resp, err := client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: openai.GPT3Dot5Turbo,
					Messages: []openai.ChatCompletionMessage{{
						Role:    openai.ChatMessageRoleUser,
						Content: gptConfig.FeedPrompt(mailMessage),
					}},
					MaxTokens: gptConfig.MaxTokens,
				},
			)

			if err != nil {
				zap.L().Fatal("failed to create completion", zap.Error(err))
			}

			if len(resp.Choices) == 0 {
				log.Fatal("no choices returned")
			}

			log.Println(resp.Choices[0].Message.Content)
		},
	}
}
