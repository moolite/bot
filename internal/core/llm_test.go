package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/pkg/tg"
	"github.com/ollama/ollama/api"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func setupTestDB(t *testing.T) {
	t.Helper()
	is := is.New(t)
	is.NoErr(db.Open(":memory:"))
	is.NoErr(db.Migrate())
	t.Cleanup(func() { db.Close() })
}

func setupMockOllama(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	os.Setenv("OLLAMA_HOST", srv.URL)
	t.Cleanup(func() {
		os.Unsetenv("OLLAMA_HOST")
		srv.Close()
	})
}

func updMention(text string) *tg.Update {
	return &tg.Update{
		UpdateID: 1,
		Message: &tg.Message{
			MessageID: 42,
			Chat:      tg.Chat{ID: -100},
			From:      &tg.User{ID: 200, IsBot: false},
			Text:      text,
			Entities: []*tg.MessageEntity{
				{Type: tg.ENTITY_MENTION, Offset: 0, Length: len(text)},
			},
		},
	}
}

func updMentionReply(text string, replyFromBot bool) *tg.Update {
	u := updMention(text)
	if replyFromBot {
		u.Message.ReplyToMessage = &tg.Message{
			MessageID: 99,
			Chat:      tg.Chat{ID: -100},
			From:      &tg.User{ID: 123, IsBot: true},
			Text:      "previous bot response",
		}
	}
	return u
}

func TestLLMCommandEmptyText(t *testing.T) {
	is := is.New(t)
	setupTestDB(t)

	ctx := context.TODO()
	b := newBot()

	u := &tg.Update{
		UpdateID: 1,
		Message: &tg.Message{
			MessageID: 42,
			Chat:      tg.Chat{ID: -100},
			From:      &tg.User{ID: 200},
			Text:      "",
		},
	}

	s, err := LLMCommand(ctx, b, u)
	is.NoErr(err)
	is.True(s == nil)
}

func TestLLMCommandTextResponse(t *testing.T) {
	is := is.New(t)
	setupTestDB(t)
	setupMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, api.ChatResponse{
			Message: api.Message{Role: "assistant", Content: "Hello!"},
			Done:    true,
		})
	})

	ctx := context.TODO()
	b := newBot()

	s, err := LLMCommand(ctx, b, updMention("hi @bot"))
	is.NoErr(err)
	is.True(s != nil)
	is.Equal(s.Method, tg.MethodSendMessage)
	is.Equal(s.Text, "Hello!")
	is.Equal(s.ReplyToMessageID, int64(42))
	is.Equal(s.ParseMode, "html")
}

func TestLLMCommandOllamaDown(t *testing.T) {
	is := is.New(t)
	setupTestDB(t)
	setupMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"model not found"}`))
	})

	ctx := context.TODO()
	b := newBot()

	s, err := LLMCommand(ctx, b, updMention("hi @bot"))
	is.NoErr(err)
	is.True(s != nil)
	is.True(s.Text != "Hello!")
	is.Equal(s.ReplyToMessageID, int64(42))
}

func TestLLMCommandCreatesConversation(t *testing.T) {
	is := is.New(t)
	setupTestDB(t)
	setupMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, api.ChatResponse{
			Message: api.Message{Role: "assistant", Content: "ok"},
			Done:    true,
		})
	})

	ctx := context.TODO()
	b := newBot()

	_, err := LLMCommand(ctx, b, updMention("hi @bot"))
	is.NoErr(err)

	conv, err := db.GetOrCreateConversation(ctx, 200, -100)
	is.NoErr(err)
	is.True(conv != nil)
	is.True(conv.ID > 0)
}

func TestLLMCommandContinuesConversation(t *testing.T) {
	is := is.New(t)
	setupTestDB(t)

	callCount := 0
	setupMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var req api.ChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if callCount == 1 {
			is.True(len(req.Messages) >= 2)
		}
		writeJSON(w, api.ChatResponse{
			Message: api.Message{Role: "assistant", Content: "response"},
			Done:    true,
		})
	})

	ctx := context.TODO()
	b := newBot()

	s1, err := LLMCommand(ctx, b, updMention("first message"))
	is.NoErr(err)
	is.True(s1 != nil)

	s2, err := LLMCommand(ctx, b, updMentionReply("second message", true))
	is.NoErr(err)
	is.True(s2 != nil)

	conv, err := db.GetOrCreateConversation(ctx, 200, -100)
	is.NoErr(err)

	msgs, err := db.GetConversationMessages(ctx, conv.ID, 50)
	is.NoErr(err)
	is.Equal(len(msgs), 4)
	is.Equal(msgs[0].Role, "user")
	is.Equal(msgs[1].Role, "assistant")
	is.Equal(msgs[2].Role, "user")
	is.Equal(msgs[3].Role, "assistant")
}

func TestLLMCommandMessagesPersisted(t *testing.T) {
	is := is.New(t)
	setupTestDB(t)
	setupMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, api.ChatResponse{
			Message: api.Message{Role: "assistant", Content: "persisted reply"},
			Done:    true,
		})
	})

	ctx := context.TODO()
	b := newBot()

	_, err := LLMCommand(ctx, b, updMention("persist this"))
	is.NoErr(err)

	conv, err := db.GetOrCreateConversation(ctx, 200, -100)
	is.NoErr(err)

	msgs, err := db.GetConversationMessages(ctx, conv.ID, 50)
	is.NoErr(err)
	is.Equal(len(msgs), 2)
	is.Equal(msgs[0].Role, "user")
	is.Equal(msgs[0].Content, "persist this")
	is.Equal(msgs[0].TelegramMsgID, int64(42))
	is.Equal(msgs[1].Role, "assistant")
	is.Equal(msgs[1].Content, "persisted reply")
}

func TestLLMCommandCompactionTrigger(t *testing.T) {
	is := is.New(t)
	setupTestDB(t)

	setupMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, api.ChatResponse{
			Message: api.Message{Role: "assistant", Content: "ok"},
			Done:    true,
		})
	})

	ctx := context.TODO()
	b := newBot()

	conv, err := db.EnsureConversation(ctx, 200, -100)
	is.NoErr(err)

	for i := 0; i < 5; i++ {
		is.NoErr(db.InsertMessage(ctx, &db.Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        "old message",
			TelegramMsgID:  int64(i),
		}))
		is.NoErr(db.InsertMessage(ctx, &db.Message{
			ConversationID: conv.ID,
			Role:           "assistant",
			Content:        "old reply",
			TelegramMsgID:  0,
		}))
	}

	_, err = LLMCommand(ctx, b, updMention("trigger compaction"))
	is.NoErr(err)

	time.Sleep(500 * time.Millisecond)

	msgs, err := db.GetConversationMessages(ctx, conv.ID, 100)
	is.NoErr(err)
	is.True(len(msgs) < 12)
}
