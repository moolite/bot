package main

import (
	"context"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/telegram"
)

const (
	FILE_DOWNLOAD_OK = iota
	FILE_DOWNLOAD_ERR
)

func fileDownloaderWorker(workerId int, filenames <-chan []string) {
	for file := range filenames {
		fileId := file[0]
		filename := file[1]

		if err := telegram.DownloadFileId(Cfg, fileId, filename); err != nil {
			slog.Error("File failed to download", "workerId", workerId, "fileId", fileId, "filename", filename, "err", err)
		} else {
			slog.Info("downloaded file", "workerId", workerId, "filename", filename)
		}
	}
}

func SyncFolder(folder string) error {
	ctx := context.Background()
	dbResults, err := db.SelectAllMedia(ctx)
	if err != nil {
		return err
	}

	slog.Debug("synchronizing files", "files", len(dbResults), "folder", folder)

	fileList, err := os.ReadDir(folder)
	if err != nil {
		return err
	}

	var fileIds [][]string
	for _, res := range dbResults {
		skip := false
		for _, f := range fileList {
			skip = strings.Contains(f.Name(), res.Data)
		}

		if !skip {
			fileExtension := "jpg"
			if res.Kind != "photo" {
				fileExtension = "m4v"
			}
			filename := path.Join(folder, res.Data+"."+fileExtension)
			fileIds = append(fileIds, []string{res.Data, filename})
		}
	}

	slog.Debug("files to download", "files", len(fileIds))

	limiter := time.Tick(100 * time.Millisecond)
	jobs := make(chan []string, len(fileList))
	for w := 1; w <= 5; w++ {
		go fileDownloaderWorker(w, jobs)
	}

	for _, job := range fileIds {
		jobs <- job
		<-limiter
	}
	close(jobs)

	return nil
}
