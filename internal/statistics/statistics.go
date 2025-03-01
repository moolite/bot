package statistics

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/utils"
)

var (
	// errors
	ErrUnknown = errors.New("unkown kind id")
)

var (
	// cache
	mux        = new(sync.Mutex)
	liveCache  = map[int64]int64{}
	done       = make(chan bool)
	knownKinds []int64
	dbCache    []*db.StatisticsKind
)

type TextTrigger struct {
	Trigger  string
	IsRegexp bool
}

func Init() (err error) {
	slog.Debug("starting Flush co-routine")
	ticker := time.NewTicker(30 * time.Second)

	dbCache, err = db.SelectStatisticKinds(context.Background())
	if err != nil {
		slog.Error("error fetching stats kind")
		return err
	}

	for _, kind := range dbCache {
		knownKinds = append(knownKinds, kind.KindID)
	}

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				ctx := context.Background()
				err := Flush(ctx)
				if err != nil {
					slog.Error("error flushing to db", "err", err)
				}
			}
		}
	}()

	return nil
}

func Stop() {
	slog.Debug("stopping Flush co-routine")
	done <- true
}

func Send(value, kind int64) error {
	mux.Lock()
	defer mux.Unlock()

	if !utils.Contains(knownKinds, kind) {
		return ErrUnknown
	}

	if _, ok := liveCache[kind]; ok {
		liveCache[kind] = liveCache[kind] + value
	} else {
		liveCache[kind] = value
	}

	return nil
}

func Flush(ctx context.Context) error {
	mux.Lock()
	defer mux.Unlock()

	for kind, value := range liveCache {
		err := db.InsertStatistics(ctx, value, kind)
		if err != nil {
			return err
		}
	}

	// cleanup
	liveCache = map[int64]int64{}

	return nil
}

func NewKind(ctx context.Context, name, trigger string, isRegexp bool) error {
	mux.Lock()
	defer mux.Unlock()

	res, err := db.InsertStatisticsKind(ctx, &db.StatisticsKind{
		Name:     name,
		Trigger:  trigger,
		IsRegexp: isRegexp,
	})
	if err != nil {
		return err
	}

	knownKinds = append(knownKinds, res)
	return nil
}

func Prometheus(ctx context.Context) (string, error) {
	results := strings.Builder{}

	items, err := db.SelectStatisticsLatest(ctx)
	if err != nil {
		return "", err
	}

	for _, item := range items {
		results.WriteString(item.Name)
		results.WriteString(":")
		results.WriteString(fmt.Sprintf("%d\n", item.Value))
	}

	return results.String(), nil
}

func ApplyTriggers(user, text string) {
	mux.Lock()
	defer mux.Unlock()

	res := 0
	for _, kind := range dbCache {
		if kind.IsRegexp {
			rx, err := regexp.Compile(kind.Trigger)
			if err != nil {
				slog.Error("trigger regexp compilation error", "kindID", kind.KindID)
				continue
			}

			if rx.MatchString(text) {
				res++
			}
		} else {
			for _, t := range strings.Split(kind.Trigger, ",") {
				if strings.Contains(text, t) {
					res++
				}
			}
		}

		if res > 0 {
			if _, ok := liveCache[kind.KindID]; ok {
				liveCache[kind.KindID]++
			} else {
				liveCache[kind.KindID] = 1
			}
		}
	}
}
