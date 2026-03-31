package vectorstore

import (
	"encoding/gob"
	"math"
	"os"
	"sort"
	"sync"
)

type Document struct {
	ID        int
	Content   string
	Embedding []float32
}

type Store struct {
	mu   sync.RWMutex
	Docs []Document
	next int
}

func New() *Store {
	return &Store{}
}

// Add inserts a document with its embedding.
func (s *Store) Add(content string, embedding []float32) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	s.Docs = append(s.Docs, Document{
		ID:        s.next,
		Content:   content,
		Embedding: embedding,
	})
	return s.next
}

// Search returns the top-k most similar documents to the query embedding.
func (s *Store) Search(query []float32, topK int) []Document {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type scored struct {
		doc   Document
		score float32
	}

	results := make([]scored, 0, len(s.Docs))
	for _, doc := range s.Docs {
		results = append(results, scored{doc, cosineSimilarity(query, doc.Embedding)})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score // higher = more similar
	})

	if topK > len(results) {
		topK = len(results)
	}
	out := make([]Document, topK)
	for i := range out {
		out[i] = results[i].doc
	}
	return out
}

// Save serializes the store to disk using gob.
func (s *Store) Save(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(s.Docs)
}

// Load deserializes a store from disk.
func Load(path string) (*Store, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var docs []Document
	if err := gob.NewDecoder(f).Decode(&docs); err != nil {
		return nil, err
	}

	maxID := 0
	for _, d := range docs {
		if d.ID > maxID {
			maxID = d.ID
		}
	}
	return &Store{Docs: docs, next: maxID}, nil
}

func cosineSimilarity(a, b []float32) float32 {
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}
