//go:build integration

package llm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/ollama/ollama/api"
)

func ensureModel(t *testing.T, model string) {
	t.Helper()
	is := is.New(t)

	if model == "" {
		model = DefaultModel
	}

	client, err := api.ClientFromEnvironment()
	is.NoErr(err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stream := false
	err = client.Pull(ctx, &api.PullRequest{
		Model:  model,
		Stream: &stream,
	}, func(resp api.ProgressResponse) error {
		if resp.Status == "success" {
			fmt.Printf("model %s already available\n", model)
		}
		return nil
	})
	is.NoErr(err)
}

func TestIntegrationChatBasic(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	resp, err := gen.Chat(ctx, []Message{{Role: "user", Content: "Say exactly: pong"}}, nil)
	is.NoErr(err)
	is.True(resp != nil)
	is.True(resp.Content != "")
	t.Logf("response: %s", resp.Content)
}

func TestIntegrationChatToolRollDice(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	resp, err := gen.Chat(ctx, []Message{
		{Role: "system", Content: SystemPrompt},
		{Role: "user", Content: "Roll 2d6 for me please"},
	}, AllTools())
	is.NoErr(err)
	is.True(resp != nil)
	is.True(resp.Content != "")
	t.Logf("response: %s", resp.Content)
}

func TestIntegrationChatConversation(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	messages := []Message{
		{Role: "system", Content: SystemPrompt},
		{Role: "user", Content: "My favorite color is blue. Remember that."},
	}

	resp1, err := gen.Chat(ctx, messages, nil)
	is.NoErr(err)
	is.True(resp1 != nil)
	t.Logf("response 1: %s", resp1.Content)

	messages = append(messages, Message{Role: "assistant", Content: resp1.Content})
	messages = append(messages, Message{Role: "user", Content: "What is my favorite color?"})

	resp2, err := gen.Chat(ctx, messages, nil)
	is.NoErr(err)
	is.True(resp2 != nil)
	is.True(strings.Contains(strings.ToLower(resp2.Content), "blue"))
	t.Logf("response 2: %s", resp2.Content)
}

func TestIntegrationChatNoResponse(t *testing.T) {
	t.Skip("search_media requires DB connection - tested in db package")
}

func TestIntegrationSetModel(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	result := gen.SetModel(DefaultModel)
	is.Equal(result, gen)
	is.Equal(gen.model, DefaultModel)

	resp, err := gen.Chat(ctx, []Message{{Role: "user", Content: "Say hello"}}, nil)
	is.NoErr(err)
	is.True(resp != nil)
}

func TestIntegrationChatEmptyTools(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	resp, err := gen.Chat(ctx, []Message{
		{Role: "system", Content: SystemPrompt},
		{Role: "user", Content: "Say exactly: pong"},
	}, []Tool{})
	is.NoErr(err)
	is.True(resp != nil)
	is.True(resp.Content != "")
	t.Logf("response: %s", resp.Content)
}

func TestIntegrationChatContextTimeout(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	_, err = gen.Chat(ctx, []Message{
		{Role: "user", Content: "Write a long story about a dragon"},
	}, nil)

	elapsed := time.Since(start)
	t.Logf("elapsed: %v, error: %v", elapsed, err)

	is.True(err != nil)
	is.True(elapsed < 1*time.Second)
}

func TestIntegrationChatEmptyToolsConversation(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	messages := []Message{
		{Role: "system", Content: SystemPrompt},
		{Role: "user", Content: "My name is Alice. Remember that."},
	}

	resp1, err := gen.Chat(ctx, messages, []Tool{})
	is.NoErr(err)
	is.True(resp1 != nil)
	t.Logf("response 1: %s", resp1.Content)

	messages = append(messages, Message{Role: "assistant", Content: resp1.Content})
	messages = append(messages, Message{Role: "user", Content: "What is my name?"})

	resp2, err := gen.Chat(ctx, messages, []Tool{})
	is.NoErr(err)
	is.True(resp2 != nil)
	is.True(strings.Contains(strings.ToLower(resp2.Content), "alice"))
	t.Logf("response 2: %s", resp2.Content)
}

func TestIntegrationToolCallRollDiceGood(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	resp, err := gen.Chat(ctx, []Message{
		{Role: "system", Content: SystemPrompt},
		{Role: "user", Content: "Please roll a d20 for me"},
	}, AllTools())
	is.NoErr(err)
	is.True(resp != nil)
	is.True(resp.Content != "")
	is.True(strings.Contains(resp.Content, "d20") || strings.Contains(resp.Content, "20"))
	t.Logf("response: %s", resp.Content)
}

func TestIntegrationToolCallRollDiceMultiple(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	resp, err := gen.Chat(ctx, []Message{
		{Role: "system", Content: SystemPrompt},
		{Role: "user", Content: "Roll 2d6+3 and also roll 1d20"},
	}, AllTools())
	is.NoErr(err)
	is.True(resp != nil)
	is.True(resp.Content != "")
	t.Logf("response: %s", resp.Content)
}

func TestIntegrationToolCallWithBadTool(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	badTool := Tool{
		Name:        "failing_tool",
		Description: "A tool that always fails to test error handling",
		Parameters: map[string]api.ToolProperty{
			"input": {
				Type:        api.PropertyType{"string"},
				Description: "Some input",
			},
		},
		Required: []string{"input"},
		Handler: func(chatID int64, args api.ToolCallFunctionArguments) (string, error) {
			input, ok := args.Get("input")
			if !ok {
				slog.Error("failing_tool: missing input argument", "chatID", chatID, "args", args)
				return "", fmt.Errorf("missing input argument")
			}
			inputStr, ok := input.(string)
			if !ok {
				slog.Error("failing_tool: input must be a string", "chatID", chatID, "args", args, "type", fmt.Sprintf("%T", input))
				return "", fmt.Errorf("input must be a string")
			}
			if inputStr == "" {
				slog.Error("failing_tool: input cannot be empty", "chatID", chatID)
				return "", fmt.Errorf("input cannot be empty")
			}
			return "processed: " + inputStr, nil
		},
	}

	_, err = gen.Chat(ctx, []Message{
		{Role: "system", Content: "You are a helpful assistant. Use the failing_tool with an empty input string when asked."},
		{Role: "user", Content: "Use the failing tool with empty input"},
	}, []Tool{badTool})

	t.Logf("error: %v", err)
}

func TestIntegrationToolCallMissingArgsLogged(t *testing.T) {
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	toolWithMissingArgs := Tool{
		Name:        "test_missing_args",
		Description: "A tool that requires a 'value' argument. Call this tool when asked.",
		Parameters: map[string]api.ToolProperty{
			"value": {
				Type:        api.PropertyType{"string"},
				Description: "Required value parameter",
			},
		},
		Required: []string{"value"},
		Handler: func(chatID int64, args api.ToolCallFunctionArguments) (string, error) {
			value, ok := args.Get("value")
			if !ok {
				slog.Error("test_missing_args: missing required 'value' argument", "chatID", chatID, "args", args)
				return "", fmt.Errorf("missing required 'value' argument")
			}
			valueStr, ok := value.(string)
			if !ok {
				slog.Error("test_missing_args: 'value' must be a string", "chatID", chatID, "type", fmt.Sprintf("%T", value))
				return "", fmt.Errorf("'value' must be a string")
			}
			return "value received: " + valueStr, nil
		},
	}

	resp, err := gen.Chat(ctx, []Message{
		{Role: "system", Content: "You are a helpful assistant. Use the test_missing_args tool when asked to test it."},
		{Role: "user", Content: "Please use the test_missing_args tool with a valid value"},
	}, []Tool{toolWithMissingArgs})

	is.NoErr(err)
	is.True(resp != nil)
	t.Logf("response: %s", resp.Content)
}
