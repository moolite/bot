package db

import (
	"context"
	"database/sql"

	"github.com/leporo/sqlf"
	"github.com/rs/zerolog/log"
)

var (
	calloutsTable       string = "callouts"
	calloutsCreateTable string = `
CREATE TABLE IF NOT EXISTS callouts
( callout VARCHAR(128) NOT NULL
, gid     VARCHAR(64)  NOT NULL
, text    TEXT
, PRIMARY KEY(callout,gid)
, FOREIGN KEY(gid) REFERENCES groups
);`
)

type Callout struct {
	Callout string `db:"callout"`
	Text    string `db:"text"`
	GID     string `db:"gid"`
}

func (c *Callout) Clone() *Callout {
	return &Callout{
		Callout: c.Callout,
		Text:    c.Text,
		GID:     c.GID,
	}
}

func InsertCallout(ctx context.Context, c *Callout) error {
	q := sqlf.
		InsertInto(calloutsTable).
		Set("gid", c.GID).
		Set("callout", c.Callout).
		Set("text", c.Text).
		Clause(
			"ON CONFLICT(callout,gid) DO UPDATE SET text = callouts.text")

	log.Debug().
		Str("stmt", q.String()).
		Interface("args", q.Args()).
		Msg("statement")

	res, err := q.ExecAndClose(ctx, dbc)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrInsert
	}

	return nil
}

func SelectOneCallout(ctx context.Context, c *Callout) error {
	q := sqlf.
		From(calloutsTable).
		Select("callout").To(&c.Callout).
		Select("text").To(&c.Text).
		Where("callout = ?", c.Callout).
		Limit(1)

	return q.QueryRowAndClose(ctx, dbc)
}

func SelectAllCallouts(ctx context.Context, gid string) ([]string, error) {
	var callouts []string
	var callout string

	q := sqlf.
		From(calloutsTable).
		Select("callout").To(&callout).
		Where("gid = ?", gid)

	err := q.QueryAndClose(ctx, dbc, func(rows *sql.Rows) {
		callouts = append(callouts, callout)
	})
	if err != nil {
		return nil, err
	}

	return callouts, nil
}

func DeleleOneCallout(ctx context.Context, c *Callout) error {
	q := sqlf.
		DeleteFrom(calloutsTable).
		Where("gid = ? AND callout = ?", c.GID, c.Callout).
		Limit(1)

	res, err := q.ExecAndClose(ctx, dbc)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrDelete
	}
	return nil
}
