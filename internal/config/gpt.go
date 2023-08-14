package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"strings"
)

type GPT struct {

	// OpenAIKey is a secret key for performing authentication. You can get it from OpenAI.
	//
	// https://platform.openai.com/account/api-keys
	OpenAIKey string `env:"OPENAI_KEY" env-description:"OpenAI API Key" env-required:"true"`

	// MaxTokens is the maximum number of tokens to generate. Requests can use up to 2048 tokens shared between prompt and completion.
	// (One token is roughly 4 characters for normal English text)
	//
	// https://platform.openai.com/docs/api-reference/completions/create#completions/create-max_tokens
	MaxTokens int `env:"MAX_TOKENS" env-description:"Max tokens to use for completion" env-default:"70"`

	// FeedPrompts is a some kind of prompt to add before, after you real message.
	FeedPrompts struct {

		// BeforeMsg is a prompt to add before your message.
		BeforeMsg string `env:"FEED_PROMPTS_BEFORE_MSG" env-default:"pretend you are an internship message parser, I have a response from the internship program:"`

		// AfterMsg is a prompt to add after your message.
		AfterMsg string `env:"FEED_PROMPTS_AFTER_MSG" env-default:"create a summary of the answer for the following points:\nthe company, vacancy, the verdict (reject, offer, test task, etc.), reason; set an emoji for a verdict."`

		// ResponseExample is an example of a response from the internship program.
		ResponseExample string `env:"FEED_PROMPTS_RESPONSE_EXAMPLE" env-default:"Here is response example: 🏢 Company: TikTok\n📝 Vacancy: Software Engineer Working Student, 2023 start\n⛔ Verdict: Reject\n🔎 Reason: Not progressing the application at this time"`
	}
}

func (c *GPT) FeedPrompt(prompt string) string {
	return strings.Join([]string{
		c.FeedPrompts.BeforeMsg, prompt, c.FeedPrompts.AfterMsg, c.FeedPrompts.ResponseExample,
	}, "\n")
}

func NewGPT() (*GPT, error) {
	var c GPT
	if err := cleanenv.ReadEnv(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
