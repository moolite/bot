package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/dicer"
	"github.com/ollama/ollama/api"
)

func AllTools() []Tool {
	return []Tool{
		{
			Name:        "roll_dice",
			Description: "Roll dice. Pass the dice notation like '2d6+3' or just 'd20'",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"dice": map[string]any{
						"type":        "string",
						"description": "Dice notation like '2d6+3'",
					},
				},
				"required": []string{"dice"},
			},
			Handler: func(chatID int64, args api.ToolCallFunctionArguments) (string, error) {
				diceStr, ok := args.Get("dice")
				if !ok {
					return "", fmt.Errorf("missing dice argument")
				}
				diceStrStr, ok := diceStr.(string)
				if !ok {
					return "", fmt.Errorf("dice argument must be a string")
				}
				dice := dicer.New(diceStrStr)
				if len(dice) == 0 {
					return "", fmt.Errorf("no valid dice found in: %s", diceStrStr)
				}
				var sb strings.Builder
				for _, d := range dice {
					sb.WriteString(d.HTML())
				}
				return sb.String(), nil
			},
		},
		{
			Name:        "search_media",
			Description: "Search for media files by description or keyword",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search query for media",
					},
				},
				"required": []string{"query"},
			},
			Handler: func(chatID int64, args api.ToolCallFunctionArguments) (string, error) {
				query, ok := args.Get("query")
				if !ok {
					return "", fmt.Errorf("missing query argument")
				}
				queryStr, ok := query.(string)
				if !ok {
					return "", fmt.Errorf("query argument must be a string")
				}
				results, err := db.SearchMedia(context.Background(), chatID, queryStr, 0)
				if err != nil {
					return "", err
				}
				if len(results) == 0 {
					return "No media found", nil
				}
				maxResults := 5
				if len(results) < maxResults {
					maxResults = len(results)
				}
				var sb strings.Builder
				sb.WriteString("Found media:\n")
				for i := 0; i < maxResults; i++ {
					m := results[i]
					desc := m.Description
					if len(desc) > 50 {
						desc = desc[:47] + "..."
					}
					sb.WriteString(fmt.Sprintf("- %s (%s)\n", desc, m.Kind))
				}
				return sb.String(), nil
			},
		},
	}
}
