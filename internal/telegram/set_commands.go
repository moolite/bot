package telegram

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type BotCommandScope struct {
	Type string `json:"type"`
}

type SetCommandsOpts struct {
	Commands []BotCommand `json:"commands"`
	Scope    string       `json:"scope,omitempty"`
}

func SetCommands(token string, commands []BotCommand) error {
	body, err := json.Marshal(commands)
	if err != nil {
		return err
	}
	res, err := apiRequest(token, `setMyCommands`, body)
	if err != nil {
		return err
	}
	slog.Debug("setMyCommands output", "data", fmt.Sprint(res))
	return nil
}
