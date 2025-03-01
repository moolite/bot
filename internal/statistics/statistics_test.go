package statistics

import (
	"context"
	"testing"

	"github.com/matryer/is"
	"github.com/moolite/bot/internal/db"
)

func TestInit(t *testing.T) {
	is := is.New(t)
	err := db.Open(":memory:")
	is.NoErr(err)
	defer db.Close()

	err = db.Migrate()
	is.NoErr(err)

	testKinds := []*db.StatisticsKind{
		{
			Name:     "JustFoo",
			Trigger:  "foo",
			IsRegexp: false,
		},
		{
			Name:     "SimplyGoo",
			Trigger:  "foo",
			IsRegexp: false,
		},
	}

	for _, item := range testKinds {
		id, err := db.InsertStatisticsKind(context.Background(), item)
		is.NoErr(err)
		item.KindID = id
	}

	err = Init()
	is.NoErr(err)
	Stop()

	is.Equal(len(knownKinds), 2)
	is.Equal(len(dbCache), 2)
	for _, item := range testKinds {
		var ctrl *db.StatisticsKind
		for _, dbitem := range dbCache {
			if item.KindID == dbitem.KindID {
				ctrl = dbitem
			}
		}

		is.True(ctrl != nil)
		is.True(ctrl.Name == item.Name)
		is.True(ctrl.IsRegexp == item.IsRegexp)
	}

	text := "hello world, foo"
	ApplyTriggers("@username", text)
	is.True(liveCache[1] == 1)
	is.True(liveCache[2] == 1)

	Flush(context.TODO())
	res, err := db.SelectStatisticsLatest(context.TODO())
	is.NoErr(err)
	for _, i := range res {
		is.True(i.Value == 1)
	}

	err = NewKind(context.TODO(), "list", "moo,goo,boo", false)
	is.NoErr(err)

	ApplyTriggers("@username", "this is a moo, goo and evena a boo")
	res, err = db.SelectStatisticsLatest(context.TODO())
	for _, r := range res {
		if r.Name == "list" {
			is.True(r.Value == 3)
			break
		}
	}
}
