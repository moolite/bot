package db

import (
	"context"
	"time"
)

type Conversation struct {
	ID        int64     `db:"id"`
	UID       int64     `db:"uid"`
	GID       int64     `db:"gid"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Message struct {
	ID             int64     `db:"id"`
	ConversationID int64     `db:"conversation_id"`
	Role           string    `db:"role"`
	Content        string    `db:"content"`
	TelegramMsgID  int64     `db:"telegram_msg_id"`
	CreatedAt      time.Time `db:"created_at"`
}

func EnsureConversation(ctx context.Context, uid, gid int64) (*Conversation, error) {
	q, err := client.prepareStmt(
		`INSERT OR IGNORE INTO llm_conversations (uid, gid) VALUES (?, ?)`,
	)
	if err != nil {
		return nil, err
	}

	_, err = q.ExecContext(ctx, uid, gid)
	if err != nil {
		return nil, err
	}

	return GetOrCreateConversation(ctx, uid, gid)
}

func GetOrCreateConversation(ctx context.Context, uid, gid int64) (*Conversation, error) {
	q, err := client.prepareStmt(
		`SELECT id, uid, gid, created_at, updated_at FROM llm_conversations WHERE uid=? AND gid=?`,
	)
	if err != nil {
		return nil, err
	}

	var conv Conversation
	err = q.GetContext(ctx, &conv, uid, gid)
	if err != nil {
		return nil, err
	}

	return &conv, nil
}

func InsertMessage(ctx context.Context, msg *Message) error {
	q, err := client.prepareStmt(
		`INSERT INTO llm_messages (conversation_id, role, content, telegram_msg_id) VALUES (?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, msg.ConversationID, msg.Role, msg.Content, msg.TelegramMsgID)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	msg.ID = id

	return nil
}

func GetConversationMessages(ctx context.Context, convID int64, limit int) ([]Message, error) {
	q, err := client.prepareStmt(
		`SELECT id, conversation_id, role, content, telegram_msg_id, created_at
FROM llm_messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
		LIMIT ?`,
	)
	if err != nil {
		return nil, err
	}

	var messages []Message
	return messages, q.SelectContext(ctx, &messages, convID, limit)
}

func DeleteConversation(ctx context.Context, convID int64) error {
	q, err := client.prepareStmt(
		`DELETE FROM llm_conversations WHERE id = ?`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, convID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrDelete
	}

	return nil
}

func CompactConversation(ctx context.Context, convID int64, systemSummary string) error {
	tx, err := client.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM llm_messages WHERE conversation_id = ?`, convID)
	if err != nil {
		return err
	}

	if systemSummary != "" {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO llm_messages (conversation_id, role, content, telegram_msg_id) VALUES (?, 'system', ?, 0)`,
			convID, systemSummary)
		if err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE llm_conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		convID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
