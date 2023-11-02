package telegram

import (
	"bytes"
	"io"
	"net/http"
)

var (
	tgBaseApi string = "https://api.telegram.org/bot"
)

func apiRequest(token string, body []byte) ([]byte, error) {
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequest(
		"POST",
		tgBaseApi+token,
		bodyReader,
	)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	body, err = io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
