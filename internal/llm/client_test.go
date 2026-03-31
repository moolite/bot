package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/ollama/ollama/api"
)

type mockHandler func(w http.ResponseWriter, r *http.Request)

func newMockClient(t *testing.T, handler mockHandler) (*Generator, func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(handler))

	u, _ := url.Parse(srv.URL)
	client := api.NewClient(u, srv.Client())

	os.Setenv("OLLAMA_HOST", srv.URL)
	t.Cleanup(func() {
		os.Unsetenv("OLLAMA_HOST")
		srv.Close()
	})

	return &Generator{
		client: client,
		model:  DefaultModel,
		chatID: -100,
	}, srv.Close
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func TestChatTextOnly(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	gen, _ := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, api.ChatResponse{
			Message: api.Message{Role: "assistant", Content: "hello!"},
			Done:    true,
		})
	})

	resp, err := gen.Chat(ctx, []Message{{Role: "user", Content: "hi"}}, nil)
	is.NoErr(err)
	is.True(resp != nil)
	is.Equal(resp.Role, "assistant")
	is.Equal(resp.Content, "hello!")
}

func TestChatWithToolCall(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	callCount := 0
	gen, _ := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			args := api.NewToolCallFunctionArguments()
			args.Set("dice", "2d6")
			writeJSON(w, api.ChatResponse{
				Message: api.Message{
					Role:      "assistant",
					ToolCalls: []api.ToolCall{{Function: api.ToolCallFunction{Name: "roll_dice", Arguments: args}}},
				},
				Done: true,
			})
		} else {
			writeJSON(w, api.ChatResponse{
				Message: api.Message{Role: "assistant", Content: "You rolled 2d6!"},
				Done:    true,
			})
		}
	})

	resp, err := gen.Chat(ctx, []Message{{Role: "user", Content: "roll 2d6"}}, AllTools())
	is.NoErr(err)
	is.True(resp != nil)
	is.Equal(resp.Content, "You rolled 2d6!")
	is.Equal(callCount, 2)
}

func TestChatMaxIterations(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	callCount := 0
	gen, _ := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		args := api.NewToolCallFunctionArguments()
		args.Set("dice", "1d6")
		writeJSON(w, api.ChatResponse{
			Message: api.Message{
				Role:      "assistant",
				ToolCalls: []api.ToolCall{{Function: api.ToolCallFunction{Name: "roll_dice", Arguments: args}}},
			},
			Done: true,
		})
	})

	_, err := gen.Chat(ctx, []Message{{Role: "user", Content: "roll forever"}}, AllTools())
	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "exceeded max tool iterations"))
	is.Equal(callCount, maxToolIterations)
}

func TestChatUnknownToolSkipped(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	callCount := 0
	gen, _ := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			args := api.NewToolCallFunctionArguments()
			writeJSON(w, api.ChatResponse{
				Message: api.Message{
					Role:      "assistant",
					ToolCalls: []api.ToolCall{{Function: api.ToolCallFunction{Name: "nonexistent_tool", Arguments: args}}},
				},
				Done: true,
			})
		} else {
			writeJSON(w, api.ChatResponse{
				Message: api.Message{Role: "assistant", Content: "I don't have that tool"},
				Done:    true,
			})
		}
	})

	resp, err := gen.Chat(ctx, []Message{{Role: "user", Content: "use unknown"}}, AllTools())
	is.NoErr(err)
	is.True(resp != nil)
	is.Equal(resp.Content, "I don't have that tool")
}

func TestChatToolHandlerError(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	gen, _ := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		args := api.NewToolCallFunctionArguments()
		writeJSON(w, api.ChatResponse{
			Message: api.Message{
				Role:      "assistant",
				ToolCalls: []api.ToolCall{{Function: api.ToolCallFunction{Name: "roll_dice", Arguments: args}}},
			},
			Done: true,
		})
	})

	brokenTool := Tool{
		Name:        "roll_dice",
		Description: "broken",
		Handler: func(chatID int64, args api.ToolCallFunctionArguments) (string, error) {
			return "", fmt.Errorf("tool exploded")
		},
	}

	_, err := gen.Chat(ctx, []Message{{Role: "user", Content: "roll"}}, []Tool{brokenTool})
	is.True(err != nil)
	is.Equal(err.Error(), "tool exploded")
}

func TestChatNoToolsProvided(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	gen, _ := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, api.ChatResponse{
			Message: api.Message{Role: "assistant", Content: "plain response"},
			Done:    true,
		})
	})

	resp, err := gen.Chat(ctx, []Message{{Role: "user", Content: "hi"}}, nil)
	is.NoErr(err)
	is.Equal(resp.Content, "plain response")
}

func TestChatMultipleToolCallsInOneResponse(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	callCount := 0
	gen, _ := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			args1 := api.NewToolCallFunctionArguments()
			args1.Set("dice", "1d6")
			args2 := api.NewToolCallFunctionArguments()
			args2.Set("dice", "1d20")
			writeJSON(w, api.ChatResponse{
				Message: api.Message{
					Role: "assistant",
					ToolCalls: []api.ToolCall{
						{Function: api.ToolCallFunction{Name: "roll_dice", Arguments: args1}},
						{Function: api.ToolCallFunction{Name: "roll_dice", Arguments: args2}},
					},
				},
				Done: true,
			})
		} else {
			writeJSON(w, api.ChatResponse{
				Message: api.Message{Role: "assistant", Content: "rolled both"},
				Done:    true,
			})
		}
	})

	resp, err := gen.Chat(ctx, []Message{{Role: "user", Content: "roll d6 and d20"}}, AllTools())
	is.NoErr(err)
	is.Equal(resp.Content, "rolled both")
	is.Equal(callCount, 2)
}

func TestChatContextCancellation(t *testing.T) {
	is := is.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	gen, _ := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		writeJSON(w, api.ChatResponse{
			Message: api.Message{Role: "assistant", Content: "too late"},
			Done:    true,
		})
	})

	_, err := gen.Chat(ctx, []Message{{Role: "user", Content: "hi"}}, nil)
	is.True(err != nil)
}
