/*
Package vectorstore provides an in-memory vector similarity search store.

It stores documents with associated float32 embeddings and supports
cosine-similarity-based retrieval of the top-k most similar documents.

# Basic Usage

Create a new store and add documents with their precomputed embeddings:

	store := vectorstore.New()
	id := store.Add("hello world", []float32{0.1, 0.2, 0.3})

Search for similar documents by providing a query embedding:

	results := store.Search([]float32{0.1, 0.2, 0.3}, 5)

# Persistence

Save and load the store to/from disk using gob encoding:

	err := store.Save("vectors.gob")
	loaded, err := vectorstore.Load("vectors.gob")

# Thread Safety

All public methods are safe for concurrent use. Add acquires a write lock,
while Search, and Save acquire read locks.
*/
package vectorstore
