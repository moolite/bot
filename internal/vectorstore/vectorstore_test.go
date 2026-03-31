package vectorstore

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/matryer/is"
)

func TestNew(t *testing.T) {
	is := is.New(t)

	s := New()
	is.Equal(s.Docs, nil)
	is.Equal(s.next, 0)
}

func TestAdd(t *testing.T) {
	is := is.New(t)

	s := New()

	tests := []struct {
		content   string
		embedding []float32
		wantID    int
	}{
		{"first", []float32{1, 0, 0}, 1},
		{"second", []float32{0, 1, 0}, 2},
		{"third", []float32{0, 0, 1}, 3},
	}

	for _, tc := range tests {
		t.Run(tc.content, func(t *testing.T) {
			is := is.New(t)
			got := s.Add(tc.content, tc.embedding)
			is.Equal(got, tc.wantID)
		})
	}

	is.Equal(len(s.Docs), 3)
	is.Equal(s.next, 3)
}

func TestSearch(t *testing.T) {
	is := is.New(t)

	s := New()
	s.Add("x-axis", []float32{1, 0, 0})
	s.Add("y-axis", []float32{0, 1, 0})
	s.Add("xy-diag", []float32{1, 1, 0})

	tests := []struct {
		name    string
		query   []float32
		topK    int
		wantIDs []int
	}{
		{
			name:    "exact match x-axis",
			query:   []float32{1, 0, 0},
			topK:    1,
			wantIDs: []int{1},
		},
		{
			name:    "closest to y-axis",
			query:   []float32{0, 1, 0},
			topK:    2,
			wantIDs: []int{2, 3},
		},
		{
			name:    "closest to xy-diagonal",
			query:   []float32{1, 1, 0},
			topK:    3,
			wantIDs: []int{3, 1, 2},
		},
		{
			name:    "topK exceeds doc count",
			query:   []float32{1, 0, 0},
			topK:    10,
			wantIDs: []int{1, 3, 2},
		},
		{
			name:    "topK zero",
			query:   []float32{1, 0, 0},
			topK:    0,
			wantIDs: []int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			got := s.Search(tc.query, tc.topK)

			gotIDs := make([]int, len(got))
			for i, d := range got {
				gotIDs[i] = d.ID
			}
			is.Equal(gotIDs, tc.wantIDs)
		})
	}
}

func TestSearch_EmptyStore(t *testing.T) {
	is := is.New(t)

	s := New()
	results := s.Search([]float32{1, 0, 0}, 5)
	is.Equal(len(results), 0)
}

func TestSearch_ZeroVector(t *testing.T) {
	is := is.New(t)

	s := New()
	s.Add("zero", []float32{0, 0, 0})
	s.Add("unit", []float32{1, 0, 0})

	results := s.Search([]float32{0, 0, 0}, 2)
	is.Equal(len(results), 2)
}

func TestSaveLoad(t *testing.T) {
	is := is.New(t)

	s := New()
	s.Add("hello", []float32{1, 2, 3})
	s.Add("world", []float32{4, 5, 6})

	dir := t.TempDir()
	path := filepath.Join(dir, "test.gob")

	is.NoErr(s.Save(path))

	loaded, err := Load(path)
	is.NoErr(err)

	is.Equal(len(loaded.Docs), 2)
	is.Equal(loaded.Docs[0].Content, "hello")
	is.Equal(loaded.Docs[0].Embedding, []float32{1, 2, 3})
	is.Equal(loaded.Docs[1].Content, "world")
	is.Equal(loaded.next, 2)

	nextID := loaded.Add("after load", []float32{7, 8, 9})
	is.Equal(nextID, 3)
}

func TestLoad_NotFound(t *testing.T) {
	is := is.New(t)

	_, err := Load("/nonexistent/path/store.gob")
	is.True(err != nil)
}

func TestSave_BadPath(t *testing.T) {
	is := is.New(t)

	s := New()
	s.Add("doc", []float32{1})

	err := s.Save("/nonexistent/dir/store.gob")
	is.True(err != nil)
}

func TestSaveLoad_EmptyStore(t *testing.T) {
	is := is.New(t)

	s := New()

	dir := t.TempDir()
	path := filepath.Join(dir, "empty.gob")

	is.NoErr(s.Save(path))

	loaded, err := Load(path)
	is.NoErr(err)
	is.Equal(len(loaded.Docs), 0)
	is.Equal(loaded.next, 0)
}

func TestCosineSimilarity(t *testing.T) {
	is := is.New(t)

	tests := []struct {
		name string
		a    []float32
		b    []float32
		want float64
	}{
		{"identical", []float32{1, 0, 0}, []float32{1, 0, 0}, 1.0},
		{"orthogonal", []float32{1, 0, 0}, []float32{0, 1, 0}, 0.0},
		{"opposite", []float32{1, 0, 0}, []float32{-1, 0, 0}, -1.0},
		{"zero query", []float32{0, 0, 0}, []float32{1, 0, 0}, 0.0},
		{"zero both", []float32{0, 0, 0}, []float32{0, 0, 0}, 0.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			got := cosineSimilarity(tc.a, tc.b)
			is.True(math.Abs(float64(got)-tc.want) < 1e-6)
		})
	}
}

func benchTool(b *testing.B, num, size int) (*Store, []float32) {
	b.Helper()

	s := New()
	for i := range num {
		emb := make([]float32, size)
		for j := range emb {
			emb[j] = float32(i+j) * 0.01
		}
		s.Add("doc", emb)
	}

	query := make([]float32, size)
	for i := range query {
		query[i] = 0.5
	}

	return s, query
}

func BenchmarkSearch10k128(b *testing.B) {
	s, query := benchTool(b, 100000, 128)

	for b.Loop() {
		s.Search(query, 10)
	}
}

func BenchmarkSearch100k512(b *testing.B) {
	s, query := benchTool(b, 100000, 512)

	for b.Loop() {
		s.Search(query, 10)
	}
}
