package core

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/moolite/bot/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fastjson"
)

var (
	resp404 = []byte(`404 - not found`)
)

func Listen(cfg *config.Config) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("marrano-bot"))
	})

	r.Post("/{token}", func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token != cfg.Telegram.Token {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello bot!"))
	})

	r.Post("/{token}", func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token != cfg.Telegram.Token {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("error reading body")

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		jsonParser, err := fastjson.ParseBytes(body)
		if err != nil {
			log.Error().Err(err).Msg("error parsing JSON")

			w.WriteHeader(http.StatusBadRequest)
			return
		}
		result, err := Handler(jsonParser)
		if err != nil {
			log.Error().Err(err).Msg("error producing response")

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(result)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(resp404)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r)
}
