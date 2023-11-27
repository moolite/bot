package statistics

import (
	"context"
	"testing"

	"github.com/moolite/bot/internal/db"
	"gotest.tools/assert"
)

func TestInit(t *testing.T) {
	err := db.Open(":memory:")
	assert.NilError(t, err)

	err = db.Migrate()
	assert.NilError(t, err)

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
		assert.NilError(t, err)
		item.KindID = id
	}

	err = Init()
	assert.NilError(t, err)
	Stop()

	assert.Equal(t, len(knownKinds), 2, "knownKinds should have length of 2")
	assert.Equal(t, len(dbCache), 2, "dbCache should have lenght of 2")
	for _, item := range testKinds {
		var ctrl *db.StatisticsKind
		for _, dbitem := range dbCache {
			if item.KindID == dbitem.KindID {
				ctrl = dbitem
			}
		}

		assert.Assert(t, ctrl != nil, "should find")
		assert.Assert(t, ctrl.Name == item.Name)
		assert.Assert(t, ctrl.IsRegexp == item.IsRegexp)
	}

	text := "hello world, foo"
	ApplyTriggers("gid", "@username", text)
	assert.Assert(t, liveCache[1] == 1)
	assert.Assert(t, liveCache[2] == 1)

	Flush(context.TODO())
	res, err := db.SelectStatisticsLatest(context.TODO())
	assert.NilError(t, err)
	for _, i := range res {
		assert.Assert(t, i.Value == 1)
	}

	err = NewKind(context.TODO(), "list", "moo,goo,boo", false)
	assert.NilError(t, err)

	ApplyTriggers("gid", "@username", "this is a moo, goo and evena a boo")
	res, err = db.SelectStatisticsLatest(context.TODO())
	for _, r := range res {
		if r.Name == "list" {
			assert.Assert(t, r.Value == 3)
			break
		}
	}
}
