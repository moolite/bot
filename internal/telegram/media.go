package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/moolite/bot/internal/config"
	"github.com/valyala/fastjson"
)

var (
	tgBaseFileApi string = "https://api.telegram.org/file/bot"
)

func GetLink(cfg *config.Config, id string) (string, error) {
	body := map[string]string{
		"file_id": id,
	}

	bodyJson, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := apiRequest(cfg.Telegram.Token, "getFile", bodyJson)
	if err != nil {
		return "", err
	}

	slog.Debug("api request result", "json", string(resp))

	p, err := fastjson.ParseBytes(resp)
	if err != nil {
		return "", err
	}

	if ok := p.GetBool("ok"); ok != true {
		return "", fmt.Errorf("error while fetching result: %s", resp)
	}

	fileId := string(p.GetStringBytes("result", "file_path"))
	if fileId == "" {
		return "", fmt.Errorf("file_path not found in json")
	}

	return tgBaseFileApi + cfg.Telegram.Token + "/" + fileId, nil
}

func DownloadFileId(cfg *config.Config, id, filename string) error {
	uri, err := GetLink(cfg, id)
	if err != nil {
		return err
	}

	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("404 file not found.")
	}

	if resp.StatusCode > 201 {
		return fmt.Errorf("error downloading file. StatusCode %d", resp.StatusCode)
	}

	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	if _, err := io.Copy(fd, resp.Body); err != nil {
		return err
	}
	return nil
}
