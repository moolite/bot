package db

import (
	"context"
)

type History struct {
	RowID     int64  `db:"rowid"`
	GID       string `db:"gid"`
	MessageID int    `db:"message_id"`
	UserID    int64  `db:"user_id"`
	Username  string `db:"username"`
	Text      string `db:"text"`
	Timestamp int64  `db:"timestamp"`
	Embedding []byte `db:"embedding"`
}

func InsertHistory(ctx context.Context, h *History) error {
	q, err := client.prepareStmt(
		`INSERT OR IGNORE INTO history (gid, message_id, user_id, username, text, timestamp, embedding)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, h.GID, h.MessageID, h.UserID, h.Username, h.Text, h.Timestamp, h.Embedding)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrInsert
	}

	return nil
}

func InsertHistoryBatch(ctx context.Context, batch []*History) error {
	tx, err := client.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, h := range batch {
		_, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO history (gid, message_id, user_id, username, text, timestamp, embedding)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			h.GID, h.MessageID, h.UserID, h.Username, h.Text, h.Timestamp, h.Embedding,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func SelectHistoryWithoutEmbeddings(ctx context.Context, gid string, limit int) ([]History, error) {
	results := []History{}
	q, err := client.prepareStmt(
		`SELECT rowid, gid, message_id, user_id, username, text, timestamp, embedding
		FROM history WHERE embedding IS NULL AND gid=? ORDER BY rowid LIMIT ?`,
	)
	if err != nil {
		return results, err
	}

	return results, q.SelectContext(ctx, &results, gid, limit)
}

func UpdateHistoryEmbedding(ctx context.Context, rowid int64, embedding []byte) error {
	q, err := client.prepareStmt(
		`UPDATE history SET embedding=? WHERE rowid=?`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, embedding, rowid)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrInsert
	}

	return nil
}

func UpdateHistoryEmbeddingBatch(ctx context.Context, items []struct {
	Rowid     int64
	Embedding []byte
}) error {
	tx, err := client.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		_, err := tx.ExecContext(ctx,
			`UPDATE history SET embedding=? WHERE rowid=?`,
			item.Embedding, item.Rowid,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func SelectHistoryWithEmbeddings(ctx context.Context, gid string) ([]History, error) {
	results := []History{}
	q, err := client.prepareStmt(
		`SELECT rowid, gid, message_id, user_id, username, text, timestamp, embedding
		FROM history WHERE embedding IS NOT NULL AND gid=?`,
	)
	if err != nil {
		return results, err
	}

	return results, q.SelectContext(ctx, &results, gid)
}

func CountHistory(ctx context.Context, gid string) (int, error) {
	q, err := client.prepareStmt(
		`SELECT COUNT(*) FROM history WHERE gid=?`,
	)
	if err != nil {
		return 0, err
	}

	var count int
	return count, q.GetContext(ctx, &count, gid)
}

func CountHistoryEmbedded(ctx context.Context, gid string) (int, error) {
	q, err := client.prepareStmt(
		`SELECT COUNT(*) FROM history WHERE embedding IS NOT NULL AND gid=?`,
	)
	if err != nil {
		return 0, err
	}

	var count int
	return count, q.GetContext(ctx, &count, gid)
}
