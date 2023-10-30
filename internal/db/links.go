package db

import (
	"github.com/leporo/sqlf"
)

var (
	linksTable       string = "links"
	linksCreateTable string = `
CREATE TABLE links (
	url VARCHAR 256 NOT NULL,
	text TEXT,
	GID VARCHAR 64 NOT NULL,

	PRIMARY KEY(url,gid),
	FOREIGN KEY(gid) REFERENCES groups,
)
`
)

type Links struct {
	URL  string
	Text string
	GID  string
}

func (l *Links) Clone() *Links {
	return &Links{
		URL:  l.URL,
		Text: l.Text,
		GID:  l.GID,
	}
}

func (l *Links) Random() *sqlf.Stmt {
	return getRandom(linksTable, l.GID).
		Bind(l)
}

func (l *Links) One() *sqlf.Stmt {
	return sqlf.
		Select("url", &l.URL).
		Select("text", &l.Text).
		Select("gid", &l.GID).
		From(linksTable).
		Where("gid = ?", l.GID)
}

func (l *Links) Insert() *sqlf.Stmt {
	return sqlf.
		InsertInto(linksTable).
		Set("url", l.URL).
		Set("text", l.Text).
		Set("gid", l.GID).
		Clause("ON CONFLICT url,gid DO UPDATE SET text = links.text")
}

func (l *Links) Delete() *sqlf.Stmt {
	return sqlf.
		DeleteFrom(linksTable).
		Where("url = ? AND gid = ?", l.URL, l.GID)
}
