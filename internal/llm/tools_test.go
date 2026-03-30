package llm

import (
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/ollama/ollama/api"
)

func makeArgs(kv ...string) api.ToolCallFunctionArguments {
	args := api.NewToolCallFunctionArguments()
	for i := 0; i < len(kv); i += 2 {
		args.Set(kv[i], kv[i+1])
	}
	return args
}

func TestAllTools(t *testing.T) {
	is := is.New(t)

	tools := AllTools()
	is.Equal(len(tools), 2)

	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
		is.True(tool.Description != "")
		is.True(tool.Handler != nil)
		is.True(tool.Parameters != nil)
	}
	is.True(names["roll_dice"])
	is.True(names["search_media"])
}

func TestRollDiceTool(t *testing.T) {
	is := is.New(t)

	tests := []struct {
		name       string
		args       api.ToolCallFunctionArguments
		wantErr    bool
		wantSubstr string
		skipExec   bool
	}{
		{
			name:       "valid 1d20",
			args:       makeArgs("dice", "1d20"),
			wantErr:    false,
			wantSubstr: "d20",
		},
		{
			name:       "valid 2d6+3",
			args:       makeArgs("dice", "2d6+3"),
			wantErr:    false,
			wantSubstr: "d6",
		},
		{
			name:       "valid 4d6k3",
			args:       makeArgs("dice", "4d6k3"),
			wantErr:    false,
			wantSubstr: "d6",
		},
		{
			name:       "multi dice 3d6 4d8",
			args:       makeArgs("dice", "3d6 4d8"),
			wantErr:    false,
			wantSubstr: "d6",
		},
		{
			name:     "missing dice arg",
			args:     makeArgs(),
			wantErr:  true,
			skipExec: true,
		},
		{
			name:     "wrong type",
			args:     makeArgs("dice", ""),
			wantErr:  true,
			skipExec: true,
		},
		{
			name:     "invalid notation",
			args:     makeArgs("dice", "xyz"),
			wantErr:  true,
			skipExec: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			var tool *Tool
			for _, t := range AllTools() {
				if t.Name == "roll_dice" {
					tool = &t
					break
				}
			}
			is.True(tool != nil)

			result, err := tool.Handler(123, tc.args)

			if tc.wantErr {
				is.True(err != nil)
			} else {
				is.NoErr(err)
				is.True(result != "")
				if tc.wantSubstr != "" {
					is.True(strings.Contains(result, tc.wantSubstr))
				}
			}

			if !tc.skipExec {
				is.True(strings.Contains(result, "<b>") || strings.Contains(result, "<i>"))
			}
		})
	}
}

func TestSearchMediaToolRequiresDB(t *testing.T) {
	is := is.New(t)

	var tool *Tool
	for _, t := range AllTools() {
		if t.Name == "search_media" {
			tool = &t
			break
		}
	}
	is.True(tool != nil)

	t.Run("missing query arg", func(t *testing.T) {
		is := is.New(t)
		_, err := tool.Handler(123, makeArgs())
		is.True(err != nil)
	})

	t.Run("wrong type", func(t *testing.T) {
		is := is.New(t)
		args := api.NewToolCallFunctionArguments()
		args.Set("query", 123)
		_, err := tool.Handler(123, args)
		is.True(err != nil)
	})

	t.Run("requires db connection", func(t *testing.T) {
		t.Skip("search_media requires active DB connection, tested in db package")
	})
}
