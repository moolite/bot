// Package llm is a client for ollama
package llm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/ollama/ollama/api"
)

const (
	DefaultModel   = `qwen3.5:0.8b`
	DefaultTimeout = 120 * time.Second
)

const maxToolIterations = 10

const SystemPrompt = `You are MarranoBot, an assistant in a Telegram group chat. You will need to speak italian to the users if they talk you in italian.
remember:
- 'marrano' is a good adjective use by humans
- 'umme' means thinking for humans

You have access to these tools:
- roll_dice: Roll dice using standard notation (e.g. '1d20', '2d6+3', '4d6k3'). Always include the 'dice' parameter.
- search_media: Search for media files in chat history by description or keyword. Always include the 'query' parameter.

When using tools, always include the required parameters. Keep responses concise, use the adjective marrano to compliment the user. Use HTML formatting when helpful.`

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]api.ToolProperty
	Required    []string
	Handler     func(chatID int64, args api.ToolCallFunctionArguments) (string, error)
}

type Generator struct {
	mu             sync.Mutex
	client         *api.Client
	model          string
	chatID         int64
	contextSize    int
	temperature    float32
	temperatureSet bool
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

func (g *Generator) SetContextSize(size int) *Generator {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.contextSize = size
	return g
}

func (g *Generator) SetTemperature(temp float32) *Generator {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.temperature = temp
	g.temperatureSet = true
	return g
}

func (g *Generator) Chat(ctx context.Context, messages []Message, tools []Tool) (*Message, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

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
			props.Set(k, v)
		}
		apiTools[i] = api.Tool{
			Type: "function",
			Function: api.ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters: api.ToolFunctionParameters{
					Type:       "object",
					Properties: props,
					Required:   t.Required,
				},
			},
		}
	}

	stream := true
	var options map[string]any
	if g.contextSize > 0 || g.temperatureSet {
		options = make(map[string]any)
		if g.contextSize > 0 {
			options["num_ctx"] = g.contextSize
		}
		if g.temperatureSet {
			options["temperature"] = g.temperature
		}
	}
	for i := range maxToolIterations {
		req := &api.ChatRequest{
			Model:    g.model,
			Messages: apiMessages,
			Tools:    apiTools,
			Stream:   &stream,
			Options:  options,
		}

		var finalMsg *Message
		var contentBuilder strings.Builder
		err := g.client.Chat(ctx, req, func(resp api.ChatResponse) error {
			if resp.Message.ToolCalls != nil {
				for _, tc := range resp.Message.ToolCalls {
					tool, ok := toolMap[tc.Function.Name]
					if !ok {
						slog.Debug("Chat: unknown tool", "name", tc.Function.Name)
						continue
					}
					result, err := tool.Handler(g.chatID, tc.Function.Arguments)
					if err != nil {
						return err
					}
					slog.Debug("Chat: tool executed", "name", tc.Function.Name, "result", result)
					apiMessages = append(apiMessages, api.Message{
						Role:    "tool",
						Content: result,
					})
				}
				return nil
			}

			if resp.Message.Content != "" {
				contentBuilder.WriteString(resp.Message.Content)
			}

			if resp.Done {
				finalContent := contentBuilder.String()
				if finalContent != "" {
					finalMsg = &Message{
						Role:    "assistant",
						Content: finalContent,
					}
				} else {
					slog.Debug("Chat: response done but no content", "iteration", i+1)
				}
			}
			return nil
		})
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("Chat: request timed out after %v", DefaultTimeout)
			}
			return nil, err
		}
		if finalMsg != nil {
			return finalMsg, nil
		}

		slog.Debug("Chat: tool call iteration", "iteration", i+1, "messages", len(apiMessages))
	}

	return nil, fmt.Errorf("Chat: exceeded max tool iterations (%d)", maxToolIterations)
}
