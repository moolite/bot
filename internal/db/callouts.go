package db

import (
	"context"
)

var (
	calloutsTable string = "callouts"
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
		WHERE callout=? AND gid=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	row := q.QueryRowContext(ctx, c.Callout, c.GID)
	return row.Scan(&c.GID, &c.Callout, &c.Text)
}

func SelectAllCallouts(ctx context.Context, gid string) ([]string, error) {
	var callouts []string

	q, err := prepareStmt(
		`SELECT callout FROM ` + calloutsTable + ` WHERE gid=?`,
	)
	if err != nil {
		return callouts, err
	}

	rows, err := q.QueryContext(ctx, gid)
	if err != nil {
		return callouts, err
	}
	defer rows.Close()

	for rows.Next() {
		var callout string
		err := rows.Scan(&callout)
		if err != nil {
			return callouts, err
		}

		callouts = append(callouts, callout)
	}

	return callouts, nil
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
