package embedder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ollama/ollama/api"
)

type Embedder struct {
	client    *api.Client
	model     string
	batchSize int
}

func New(model string, batchSize int) (*Embedder, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("create ollama client: %w", err)
	}
	if batchSize <= 0 {
		batchSize = 32
	}
	return &Embedder{
		client:    client,
		model:     model,
		batchSize: batchSize,
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	var all [][]float32
	for i := 0; i < len(texts); i += e.batchSize {
		end := i + e.batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]
		resp, err := e.client.Embed(ctx, &api.EmbedRequest{
			Model:    e.model,
			Input:    batch,
			Truncate: ptr(true),
		})
		if err != nil {
			return nil, fmt.Errorf("embed batch %d: %w", i/e.batchSize+1, err)
		}
		slog.Info("embed batch", "batch", i/e.batchSize+1, "count", len(batch), "total", len(texts))
		all = append(all, resp.Embeddings...)
	}
	return all, nil
}

func (e *Embedder) Model() string {
	return e.model
}

func ptr[T any](v T) *T {
	return &v
}
