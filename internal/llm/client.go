// Package llm is a client for ollama
package llm

import (
	"context"

	"github.com/ollama/ollama/api"
)

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
	Handler     func(args map[string]any) (string, error)
}

var tools map[string]Tool

func NewClient(_ int64) {
}
