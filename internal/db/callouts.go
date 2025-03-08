package db

import (
	"context"
)

var (
	calloutsTable string = "callouts"
)

type Callout struct {
	GID     int64  `db:"gid"`
	Callout string `db:"callout"`
	Text    string `db:"text"`
}

func (c *Callout) Clone() *Callout {
	return &Callout{
		Callout: c.Callout,
		Text:    c.Text,
		GID:     c.GID,
	}
}

func InsertCallout(ctx context.Context, c *Callout) error {
	q, err := prepareStmt(
		`INSERT OR REPLACE INTO ` + calloutsTable + `
		(gid,callout,text) VALUES (?,?,?)`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, c.GID, c.Callout, c.Text)
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
	q, err := prepareStmt(
		`SELECT gid,callout,text FROM ` + calloutsTable + `
		WHERE callout LIKE ? AND gid=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	return q.GetContext(ctx, c, c.Callout, c.GID)
}

func SelectAllCallouts(ctx context.Context, gid string) ([]string, error) {
	callouts := []string{}

	q, err := prepareStmt(
		`SELECT callout FROM ` + calloutsTable + ` WHERE gid=?`,
	)
	if err != nil {
		return callouts, err
	}

	return callouts, q.SelectContext(ctx, &callouts)
}

func DeleleOneCallout(ctx context.Context, c *Callout) error {
	q, err := prepareStmt(
		`DELETE FROM ` + calloutsTable + ` WHERE gid=? AND callout=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, c.GID, c.Callout)
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
