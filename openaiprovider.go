package goreact

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type LLMProvider interface {
	Request(system, prompt string) (string, error)
}

type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(openaikey string) (*OpenAIProvider, error) {
	client := openai.NewClient(openaikey)
	return &OpenAIProvider{
		client: client,
		model:  openai.GPT4o,
	}, nil
}

func (o *OpenAIProvider) WithModel(model string) *OpenAIProvider {
	o.model = model
	return o
}

func (o *OpenAIProvider) Request(system, prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:       o.model,
		Temperature: 0.1,
		Stop:        []string{"OBSERVATION:", "STOP_ACTION"},
		User:        "goreact",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: system,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := o.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
