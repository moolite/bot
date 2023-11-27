package statistics

import (
	"context"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/moolite/bot/internal/db"
)

var version string = "0.10.0"

type ChartData struct {
	Current  int64
	Previous int64
}

type TemplateData struct {
	Version string
	Title   string
	Data    map[time.Time]*ChartData
}

//go:embed plot/*
var plotFS embed.FS

func mapStatsToTemplateData(data []*db.StatisticsJoin) (res map[time.Time]*ChartData) {
	for i := 0; i < len(data); i++ {
		var prev *db.StatisticsJoin
		if i > 0 {
			prev = data[i-1]
		} else {
			prev = data[i]
		}
		curr := data[i]
		res[curr.Date] = &ChartData{
			Previous: prev.Value,
			Current:  curr.Value,
		}
	}

	return res
}

func mapChannelStatsToTemplateData(data []*db.ChannelStats) (res map[time.Time]*ChartData) {
	for i := 0; i < len(data); i++ {
		var prev *db.ChannelStats
		if i > 0 {
			prev = data[i-1]
		} else {
			prev = data[i]
		}
		curr := data[i]
		res[curr.TS] = &ChartData{
			Previous: prev.Points,
			Current:  curr.Points,
		}
	}

	return res
	return res
}

func PlotRouter() (*chi.Mux, error) {
	tmpl, err := template.ParseFS(plotFS, "plot/index.tmpl")
	if err != nil {
		slog.Error("template error", "err", err)
		return nil, err
	}

	r := chi.NewRouter()
	subfs, err := fs.Sub(plotFS, "plot")
	if err != nil {
		return nil, err
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		stats, err := db.SelectStatisticsByDateRange(ctx, time.Now(), time.Now())
		if err != nil {
			slog.Error("error selecting stats by range", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data := &TemplateData{
			Version: version,
			Title:   "marrano-bot stats",
			Data:    mapStatsToTemplateData(stats),
		}
		if err := tmpl.Execute(w, data); err != nil {
			slog.Error("error executing template", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	r.Get("/{channel}", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		channel := chi.URLParam(r, "channel")
		stats, err := db.SelectChannelStatsTimeSeries(ctx, channel)
		if err != nil {
			slog.Error("error selecting channel statistics", "err", err, "channel", channel)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data := &TemplateData{
			Version: version,
			Title:   "channel stats: " + channel,
			Data:    mapChannelStatsToTemplateData(stats),
		}
		if err := tmpl.Execute(w, data); err != nil {
			slog.Error("error executing template", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		fs := http.FileServer(http.FS(subfs))
		fs.ServeHTTP(w, r)
	})

	return r, nil
}

func statisticsJoinToChartData(channel string) map[time.Time][]*ChartData {
	var results map[time.Time][]*ChartData

	stats, err := db.SelectChannelStatsTimeSeries(context.Background(), channel)
	if err != nil {
		slog.Error("error selecting channel statistics", "err", err, "channel", channel)
		return results
	}

	for _, s := range stats {
		if _, ok := results[s.TS]; !ok {
			results[s.TS] = []*ChartData{}
		}

		results[s.TS] = append(results[s.TS], &ChartData{
			s.Points,
			s.Points,
		})
	}

	return results
}
