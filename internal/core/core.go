package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/moolite/bot/internal/config"
	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/statistics"
	"github.com/valyala/fastjson"
)

var (
	resp404 = []byte(`404 - not found`)
)

func Listen(cfg *config.Config) error {

	logger := httplog.NewLogger("marrano-bot", httplog.Options{
		JSON:     true,
		LogLevel: slog.LevelDebug,
		Concise:  true,
		// RequestHeaders:  true,
		// ResponseHeaders: true,
		MessageFieldName: "message",
		// TimeFieldFormat: time.RFC850,
		// SourceFieldName: "source",
		QuietDownRoutes: []string{
			"/",
			"/ping",
		},
		QuietDownPeriod: 10 * time.Second,
	})

	err := db.Open(cfg.Database)
	if err != nil {
		slog.Error("error opening connection", "err", err)
		return err
	}

	err = statistics.Init()
	if err != nil {
		slog.Error("error initializing statistics", "err", err)
		return err
	}
	defer statistics.Stop()

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("marrano-bot"))
	})

	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		prometheusData, err := statistics.Prometheus(context.Background())
		if err != nil {
			slog.Error("error producing prometheus statistics", "err", err)

			http.Error(w, "Error producing prometheus statistics", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "text")
		w.Write([]byte(prometheusData))
	})

	r.Get("/stats.json", func(w http.ResponseWriter, r *http.Request) {
		data, err := db.SelectStatisticsLatest(context.Background())
		if err != nil {
			slog.Error("error selecting latest statistics", "err", err)

			http.Error(w, "Error producing statistics", http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(data)
		if err != nil {
			slog.Error("error mashaling statistics", "err", err)

			http.Error(w, "error mashaling statistics", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write(resp)
	})

	r.Get("/t/{apikey}", func(w http.ResponseWriter, r *http.Request) {
		oplog := httplog.LogEntry(r.Context())
		if apikey := chi.URLParam(r, "apikey"); apikey != cfg.Telegram.ApiKey {
			w.WriteHeader(http.StatusNotFound)
			oplog.Error("apikey not found")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello bot!"))
	})

	r.Post("/t/{apikey}", func(w http.ResponseWriter, r *http.Request) {
		oplog := httplog.LogEntry(r.Context())
		if apikey := chi.URLParam(r, "apikey"); apikey != cfg.Telegram.ApiKey {
			w.WriteHeader(http.StatusNotFound)
			oplog.Error("apikey not found")
			return
		}

		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			oplog.Error("body not defined")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			oplog.Error("body read error", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		oplog.Debug("fastjson is killing me!", "body", string(body))

		jsonParser, err := fastjson.ParseBytes(body)
		if err != nil {
			oplog.Error("body parse error", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		result, err := Handler(jsonParser)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				oplog.Debug("handler returned empty response", "err", err)
				w.WriteHeader(http.StatusOK)
				return
			}

			oplog.Error("error producing response", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err = result.Marshal()
		if err != nil {
			oplog.Error("error producing response", "err", err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		oplog.Debug("bot response", "body", body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(resp404)
	})

	slog.Info("http handler listening", "port", cfg.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r)
}
