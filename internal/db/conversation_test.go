package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/matryer/is"
)

func TestConversationEnsure(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	os.WriteFile(dbPath, []byte{}, 0644)
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)
	is.True(conv != nil)
	is.True(conv.ID > 0)
	is.Equal(conv.UID, int64(100))
	is.Equal(conv.GID, int64(-200))
}

func TestConversationEnsureIdempotent(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv1, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	conv2, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	is.Equal(conv1.ID, conv2.ID)
}

func TestConversationGetOrCreateExisting(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv1, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	conv2, err := GetOrCreateConversation(ctx, 100, -200)
	is.NoErr(err)
	is.Equal(conv1.ID, conv2.ID)
}

func TestConversationGetOrCreateNotFound(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	_, err := GetOrCreateConversation(ctx, 999, -999)
	is.True(err != nil)
}

func TestConversationDifferentUsers(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv1, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	conv2, err := EnsureConversation(ctx, 200, -200)
	is.NoErr(err)

	is.True(conv1.ID != conv2.ID)
}

func TestInsertMessage(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	msg := &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "hello",
		TelegramMsgID:  42,
	}
	is.NoErr(InsertMessage(ctx, msg))
	is.True(msg.ID > 0)
}

func TestGetConversationMessages(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	msgs := []*Message{
		{ConversationID: conv.ID, Role: "user", Content: "hello", TelegramMsgID: 1},
		{ConversationID: conv.ID, Role: "assistant", Content: "hi there", TelegramMsgID: 0},
		{ConversationID: conv.ID, Role: "user", Content: "how are you?", TelegramMsgID: 2},
	}
	for _, m := range msgs {
		is.NoErr(InsertMessage(ctx, m))
	}

	result, err := GetConversationMessages(ctx, conv.ID, 50)
	is.NoErr(err)
	is.Equal(len(result), 3)
	is.Equal(result[0].Role, "user")
	is.Equal(result[0].Content, "hello")
	is.Equal(result[1].Role, "assistant")
	is.Equal(result[2].Role, "user")
}

func TestGetConversationMessagesLimit(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	for i := 0; i < 10; i++ {
		is.NoErr(InsertMessage(ctx, &Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        "msg",
			TelegramMsgID:  int64(i),
		}))
	}

	result, err := GetConversationMessages(ctx, conv.ID, 5)
	is.NoErr(err)
	is.Equal(len(result), 5)
}

func TestGetConversationMessagesEmpty(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	result, err := GetConversationMessages(ctx, conv.ID, 50)
	is.NoErr(err)
	is.Equal(len(result), 0)
}

func TestDeleteConversation(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	is.NoErr(InsertMessage(ctx, &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "hello",
		TelegramMsgID:  1,
	}))

	is.NoErr(DeleteConversation(ctx, conv.ID))

	_, err = GetOrCreateConversation(ctx, 100, -200)
	is.True(err != nil)
}

func TestDeleteConversationNotFound(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	err := DeleteConversation(ctx, 99999)
	is.Equal(err, ErrDelete)
}

func TestCompactConversation(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	for i := 0; i < 5; i++ {
		is.NoErr(InsertMessage(ctx, &Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        "some message",
			TelegramMsgID:  int64(i),
		}))
	}

	is.NoErr(CompactConversation(ctx, conv.ID, "summary of conversation"))

	result, err := GetConversationMessages(ctx, conv.ID, 50)
	is.NoErr(err)
	is.Equal(len(result), 1)
	is.Equal(result[0].Role, "system")
	is.Equal(result[0].Content, "summary of conversation")
}

func TestCompactConversationEmptySummary(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	is.NoErr(InsertMessage(ctx, &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "hello",
		TelegramMsgID:  1,
	}))

	is.NoErr(CompactConversation(ctx, conv.ID, ""))

	result, err := GetConversationMessages(ctx, conv.ID, 50)
	is.NoErr(err)
	is.Equal(len(result), 0)
}

func TestGetOrCreateSessionBySessionID(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv, err := GetOrCreateSessionBySessionID(ctx, "test-session-1", 100, -200)
	is.NoErr(err)
	is.True(conv != nil)
	is.True(conv.ID > 0)
	is.Equal(conv.SessionID, "test-session-1")
	is.Equal(conv.UID, int64(100))
	is.Equal(conv.GID, int64(-200))

	conv2, err := GetOrCreateSessionBySessionID(ctx, "test-session-1", 100, -200)
	is.NoErr(err)
	is.Equal(conv.ID, conv2.ID)
}

func TestGetOrCreateSessionBySessionIDEmpty(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	_, err := GetOrCreateSessionBySessionID(ctx, "", 100, -200)
	is.Equal(err, ErrNoSessionID)
}

func TestGetSessions(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	sessions, err := GetSessions(ctx, 10)
	is.NoErr(err)
	is.Equal(len(sessions), 0)

	conv, err := GetOrCreateSessionBySessionID(ctx, "session-1", 100, -200)
	is.NoErr(err)

	is.NoErr(InsertMessage(ctx, &Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "hello",
		TelegramMsgID:  1,
	}))

	sessions, err = GetSessions(ctx, 10)
	is.NoErr(err)
	is.Equal(len(sessions), 1)
	is.Equal(sessions[0].SessionID, "session-1")
	is.Equal(sessions[0].MsgCount, 1)
}

func TestGetSessionsFiltersEmptySessionID(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	_, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	sessions, err := GetSessions(ctx, 10)
	is.NoErr(err)
	is.Equal(len(sessions), 0)
}

func TestCountConversationMessages(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	conv, err := EnsureConversation(ctx, 100, -200)
	is.NoErr(err)

	count, err := CountConversationMessages(ctx, conv.ID)
	is.NoErr(err)
	is.Equal(count, 0)

	for i := 0; i < 5; i++ {
		is.NoErr(InsertMessage(ctx, &Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        "msg",
			TelegramMsgID:  int64(i),
		}))
	}

	count, err = CountConversationMessages(ctx, conv.ID)
	is.NoErr(err)
	is.Equal(count, 5)
}
