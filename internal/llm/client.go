package llm

import "context"

type Client interface {
	Chat(ctx context.Context, prompt string) (string, error)
	ChatStructured(ctx context.Context, prompt string, output interface{}) error
	GetModel() string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
