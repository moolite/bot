package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/embedder"
	"github.com/moolite/bot/internal/vectorstore"
)

const importBatchSize = 100

type telegramExport struct {
	ID       int64             `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Messages []telegramMessage `json:"messages"`
}

type telegramMessage struct {
	ID     int64           `json:"id"`
	Type   string          `json:"type"`
	Date   string          `json:"date"`
	From   string          `json:"from"`
	FromID string          `json:"from_id"`
	Text   json.RawMessage `json:"text"`
}

func extractText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return ""
		}
		return strings.TrimSpace(s)
	}
	var parts []any
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ""
	}
	var sb strings.Builder
	for _, p := range parts {
		switch v := p.(type) {
		case string:
			sb.WriteString(v)
		case map[string]any:
			if t, ok := v["text"].(string); ok {
				sb.WriteString(t)
			}
		}
	}
	return strings.TrimSpace(sb.String())
}

func parseTelegramTimestamp(s string) int64 {
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func ImportHistory(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var export telegramExport
	if err := json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	gid := fmt.Sprint(export.ID)
	slog.Info("parsed export", "name", export.Name, "gid", gid, "messages", len(export.Messages))

	if err := db.InsertGroup(ctx, export.ID, export.Name); err != nil {
		slog.Warn("insert group", "err", err)
	}

	var history []*db.History
	for _, msg := range export.Messages {
		if msg.Type != "message" {
			continue
		}
		text := extractText(msg.Text)
		if text == "" {
			continue
		}
		history = append(history, &db.History{
			GID:       gid,
			MessageID: int(msg.ID),
			Username:  msg.From,
			Text:      text,
			Timestamp: parseTelegramTimestamp(msg.Date),
		})
	}

	slog.Info("valid messages", "count", len(history))

	for i := 0; i < len(history); i += importBatchSize {
		end := i + importBatchSize
		if end > len(history) {
			end = len(history)
		}
		if err := db.InsertHistoryBatch(ctx, history[i:end]); err != nil {
			slog.Error("batch insert", "batch", i, "err", err)
		}
	}

	total, err := db.CountHistory(ctx, gid)
	if err != nil {
		return fmt.Errorf("count history: %w", err)
	}

	slog.Info("inserted messages", "total", total)

	emb, err := embedder.New(Cfg.LLM.EmbeddingModel, Cfg.LLM.EmbeddingBatch)
	if err != nil {
		return fmt.Errorf("create embedder: %w", err)
	}

	embedded := 0
	for {
		rows, err := db.SelectHistoryWithoutEmbeddings(ctx, gid, Cfg.LLM.EmbeddingBatch)
		if err != nil {
			return fmt.Errorf("select unembedded: %w", err)
		}
		if len(rows) == 0 {
			break
		}

		texts := make([]string, len(rows))
		for i, r := range rows {
			texts[i] = r.Text
		}

		embeddings, err := emb.Embed(ctx, texts)
		if err != nil {
			slog.Error("embedding error", "err", err)
			break
		}

		items := make([]struct {
			Rowid     int64
			Embedding []byte
		}, len(embeddings))
		for i := range embeddings {
			items[i].Rowid = rows[i].RowID
			items[i].Embedding = vectorstore.Float32ToBytes(embeddings[i])
		}

		if err := db.UpdateHistoryEmbeddingBatch(ctx, items); err != nil {
			slog.Error("update embeddings", "err", err)
		}

		embedded += len(rows)
		slog.Info("embedding progress", "embedded", embedded, "total", total)
	}

	slog.Info("import complete", "total", total, "embedded", embedded)
	return nil
}
