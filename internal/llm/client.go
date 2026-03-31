// Package llm is a client for ollama
package llm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ollama/ollama/api"
)

const (
	DefaultModel = `qwen3.5:0.5b`
)

const maxToolIterations = 10

const SystemPrompt = `You are MarranoBot, an assistant in a Telegram group chat. You have access to tools:
- roll_dice: Roll dice. Usage: describe what dice to roll (e.g., "2d6+3")
- search_media: Search for media files. Usage: search by description

Keep responses concise, use the adjective marrano to compliment the user. Use HTML formatting when helpful.`

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
	Handler     func(chatID int64, args api.ToolCallFunctionArguments) (string, error)
}

type Generator struct {
	client *api.Client
	model  string
	chatID int64
}

func NewClient(ctx context.Context, chatID int64) (*Generator, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	g := &Generator{
		client: client,
		model:  DefaultModel,
		chatID: chatID,
	}

	return g, nil
}

func (g *Generator) SetModel(model string) *Generator {
	g.model = model
	return g
}

func (g *Generator) Chat(ctx context.Context, messages []Message, tools []Tool) (*Message, error) {
	toolMap := make(map[string]Tool)
	for _, t := range tools {
		toolMap[t.Name] = t
	}

	apiMessages := make([]api.Message, len(messages))
	for i, m := range messages {
		apiMessages[i] = api.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	apiTools := make(api.Tools, len(tools))
	for i, t := range tools {
		props := api.NewToolPropertiesMap()
		for k, v := range t.Parameters {
			if prop, ok := v.(api.ToolProperty); ok {
				props.Set(k, prop)
			}
		}
		apiTools[i] = api.Tool{
			Type: "function",
			Function: api.ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters: api.ToolFunctionParameters{
					Type:       "object",
					Properties: props,
				},
			},
		}
	}

	for i := range maxToolIterations {
		req := &api.ChatRequest{
			Model:    g.model,
			Messages: apiMessages,
			Tools:    apiTools,
		}

		var finalMsg *Message
		err := g.client.Chat(ctx, req, func(resp api.ChatResponse) error {
			if resp.Message.ToolCalls != nil {
				for _, tc := range resp.Message.ToolCalls {
					tool, ok := toolMap[tc.Function.Name]
					if !ok {
						continue
					}
					result, err := tool.Handler(g.chatID, tc.Function.Arguments)
					if err != nil {
						return err
					}
					apiMessages = append(apiMessages, api.Message{
						Role:    "tool",
						Content: result,
					})
				}
				return nil
			}

			if resp.Message.Content != "" {
				finalMsg = &Message{
					Role:    "assistant",
					Content: resp.Message.Content,
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		if finalMsg != nil {
			return finalMsg, nil
		}

		slog.Debug("Chat: tool call iteration", "iteration", i+1, "messages", len(apiMessages))
	}

	return nil, fmt.Errorf("Chat: exceeded max tool iterations (%d)", maxToolIterations)
}
