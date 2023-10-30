package core

import (
	"encoding/json"

	"github.com/valyala/fastjson"
)

type TelegramMessage struct {
}

func Handler(p *fastjson.Value) ([]byte, error) {
	res := &TelegramMessage{}

	result, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return result, nil
}
