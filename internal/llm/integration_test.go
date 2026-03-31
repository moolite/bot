//go:build integration

package llm

import (
	"context"
	"fmt"
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
	is := is.New(t)
	ensureModel(t, DefaultModel)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gen, err := NewClient(ctx, -100)
	is.NoErr(err)

	resp, err := gen.Chat(ctx, []Message{
		{Role: "system", Content: SystemPrompt},
		{Role: "user", Content: "Search for media about cats"},
	}, AllTools())
	is.NoErr(err)
	is.True(resp != nil)
	t.Logf("response: %s", resp.Content)
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
