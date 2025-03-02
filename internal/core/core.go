package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"

	"github.com/mattn/go-isatty"
	_ "github.com/mattn/go-sqlite3"
	"github.com/moolite/bot/internal/config"
	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/statistics"
	"github.com/moolite/bot/pkg/tg"
)

var (
	resp404 = []byte(`404 not found`)
)

func Listen(ctx context.Context, b *tg.Bot, cfg *config.Config) error {
	logger := httplog.NewLogger("marrano-bot", httplog.Options{
		JSON:     !isatty.IsTerminal(os.Stdin.Fd()),
		LogLevel: cfg.LogLevel,
		Concise:  true,
		// RequestHeaders:  true,
		// ResponseHeaders: true,
		MessageFieldName: "message",
		// TimeFieldFormat: time.RFC850,
		// SourceFieldName: "source",
		QuietDownRoutes: []string{
			"/",
			"/ping",
			"/health",
		},
		QuietDownPeriod: 10 * time.Second,
	})

	if err := db.Open(cfg.Database); err != nil {
		slog.Error("error opening connection", "err", err)
		return err
	}

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

		b.HttpHandler(oplog)(w, r)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(resp404)
	})

	if err := registerCommands(ctx, b); err != nil {
		slog.Error("Error in registerCommands", "err", err)
		return err
	}

	// register bot event handlers
	registerBotHandlers(ctx, b)

	slog.Info("http handler listening", "port", cfg.Port)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}
	go srv.ListenAndServe()

	<-ctx.Done()

	if err := db.Close(); err != nil {
		return err
	}

	return srv.Close()
}
