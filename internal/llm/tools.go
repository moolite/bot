package llm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/dicer"
	"github.com/ollama/ollama/api"
)

func AllTools() []Tool {
	return []Tool{
		{
			Name:        "roll_dice",
			Description: "Roll dice using standard notation. Examples: '1d20' (one 20-sided die), '2d6+3' (two 6-sided dice plus 3), '4d6k3' (roll 4d6, keep highest 3).",
			Parameters: map[string]api.ToolProperty{
				"dice": {
					Type:        api.PropertyType{"string"},
					Description: "The dice notation to roll, e.g. '1d20', '2d6+3', '4d6k3'. This is required.",
				},
			},
			Required: []string{"dice"},
			Handler: func(chatID int64, args api.ToolCallFunctionArguments) (string, error) {
				diceStr, ok := args.Get("dice")
				if !ok {
					slog.Error("roll_dice: missing dice argument", "chatID", chatID, "args", args)
					return "", fmt.Errorf("missing dice argument")
				}
				diceStrStr, ok := diceStr.(string)
				if !ok {
					slog.Error("roll_dice: dice argument must be a string", "chatID", chatID, "args", args, "type", fmt.Sprintf("%T", diceStr))
					return "", fmt.Errorf("dice argument must be a string")
				}
				dice := dicer.New(diceStrStr)
				if len(dice) == 0 {
					slog.Error("roll_dice: no valid dice found", "chatID", chatID, "input", diceStrStr)
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
			Description: "Search for media files (images, videos, gifs) in the chat history by description or keyword. Returns matching media entries.",
			Parameters: map[string]api.ToolProperty{
				"query": {
					Type:        api.PropertyType{"string"},
					Description: "The search query describing what media to find, e.g. 'cats', 'funny gif', 'sunset photo'. This is required.",
				},
			},
			Required: []string{"query"},
			Handler: func(chatID int64, args api.ToolCallFunctionArguments) (string, error) {
				query, ok := args.Get("query")
				if !ok {
					slog.Error("search_media: missing query argument", "chatID", chatID, "args", args)
					return "", fmt.Errorf("missing query argument")
				}
				queryStr, ok := query.(string)
				if !ok {
					slog.Error("search_media: query argument must be a string", "chatID", chatID, "args", args, "type", fmt.Sprintf("%T", query))
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
