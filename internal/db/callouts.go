package db

import (
	"github.com/leporo/sqlf"
)

var (
	calloutsCreateTable string = `
CREATE TABLE IF NOT EXISTS callouts
(	callout VARCHAR(128) NOT NULL
,	gid     VARCHAR(64)  NOT NULL
,	PRIMARY KEY(callout,gid)
,	FOREIGN KEY(gid) REFERENCES groups
);`
)

type Callout struct {
	Callout string
	Text    string
	GID     string
}

func (c *Callout) Clone() *Callout {
	return &Callout{
		Callout: c.Callout,
		Text:    c.Text,
		GID:     c.GID,
	}
}

func InsertCallout(gid, callout, text string) *sqlf.Stmt {
	return sqlf.
		InsertInto("callouts").
		Set("callout", callout).
		Set("text", text).
		Set("gid", gid).
		Clause("ON CONFLICT callout,gid DO UPDATE SET text = callouts.text")
}

func SelectOneCallout(gid, callout string) *sqlf.Stmt {
	return sqlf.
		Select("callout", "text").
		From("callouts").
		Where("callout = ? AND gid = ?", callout, gid)
}

func SelectAllCallouts(gid string) *sqlf.Stmt {
	return sqlf.
		Select("callout").
		Where("gid = ?", gid)
}

func DeleleOneCallout(gid, callout string) *sqlf.Stmt {
	return sqlf.
		DeleteFrom("callouts").
		Where("callout = ? AND gid = ?", callout, gid)
}
