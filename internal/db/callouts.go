package db

import (
	"github.com/leporo/sqlf"
)

var (
	calloutsCreateTable string = `
CREATE OR UPDATE TABLE callouts
(
	callout VARCHAR 128 NOT NULL,
	gid     VARCHAR 64  NOT NULL,

	PRIMARY KEY(callout,gid),
	FOREIGN KEY(gid) REFERENCES groups,
)
`
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

func (c *Callout) Insert() *sqlf.Stmt {
	return sqlf.
		InsertInto("callouts").
		Set("callout", c.Callout).
		Set("text", c.Text).
		Set("gid", c.GID).
		Clause("ON CONFLICT callout,gid DO UPDATE SET text = callouts.text")
}

func (c *Callout) AllCallouts() *sqlf.Stmt {
	return sqlf.
		Select("callout").
		Where("gid = ?", c.GID)
}

func (c *Callout) One() *sqlf.Stmt {
	return sqlf.
		Select("callout", "text").
		From("callouts").
		Where("callout = ? AND gid = ?", c.Callout, c.GID).
		Bind(c)
}

func (c *Callout) DelByCallout() *sqlf.Stmt {
	return sqlf.
		DeleteFrom("callouts").
		Where("callout = ? AND gid = ?", c.Callout, c.GID)
}
