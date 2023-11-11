package db

import (
	"context"
	"database/sql"

	"github.com/leporo/sqlf"
)

var (
	groupsTable       string = "groups"
	groupsCreateTable string = `
CREATE TABLE IF NOT EXISTS groups
( gid VARCHAR(64) NOT NULL
, title text
, PRIMARY KEY(gid)
);
`
)

type Group struct {
	GID   string `db:"gid"`
	Title string `db:"title"`
}

func (g *Group) Clone() *Group {
	return &Group{
		GID:   g.GID,
		Title: g.Title,
	}
}

func SelectOneGroup(ctx context.Context, gid string) (*Group, error) {
	g := &Group{GID: gid}

	q := sqlf.
		From(groupsTable).
		Select("title", g.Title).
		Where("gid = ?", gid).
		Limit(1)

	if err := q.QueryRowAndClose(ctx, dbc); err != nil {
		return nil, err
	}

	return g, nil
}

func SelectAllGroups(ctx context.Context) ([]*Group, error) {
	var g *Group
	q := sqlf.
		From(groupsTable).
		Select("gid", g.GID).
		Select("title", g.Title)

	var ret []*Group
	err := q.QueryAndClose(ctx, dbc, func(r *sql.Rows) {
		ret = append(ret, g.Clone())
	})

	return ret, err
}

func InsertGroup(ctx context.Context, gid, title string) error {
	q := sqlf.
		InsertInto(groupsTable).
		Set("gid", gid).
		Set("title", title).
		Clause(
			"ON CONFLICT(gid) DO UPDATE SET title = groups.title")

	if res, err := q.ExecAndClose(ctx, dbc); err != nil {
		return err
	} else if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrInsert
	}

	return nil
}
