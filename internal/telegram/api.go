package telegram

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

var (
	tgBaseApi string = "https://api.telegram.org/bot"
)

func apiRequest(token, method string, body []byte) ([]byte, error) {
	bodyReader := bytes.NewBuffer(body)
	req, err := http.NewRequest(
		"POST",
		tgBaseApi+token+"/"+method,
		bodyReader,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		Timeout: 60 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return response, nil
}
