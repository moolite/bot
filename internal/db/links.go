package db

import (
	"github.com/leporo/sqlf"
)

var (
	linksTable       string = "links"
	linksCreateTable string = `
CREATE TABLE IF NOT EXISTS links
(	url VARCHAR(256) NOT NULL
,	text TEXT
,	gid VARCHAR(64) NOT NULL
,	PRIMARY KEY(url,gid)
,	FOREIGN KEY(gid) REFERENCES groups
);`
)

type Link struct {
	URL  string `db:"url"`
	Text string `db:"text"`
	GID  string `db:"gid"`
}

func (l *Link) Clone() *Link {
	return &Link{
		URL:  l.URL,
		Text: l.Text,
		GID:  l.GID,
	}
}

func SelectRandomLink() *sqlf.Stmt {
	l := new(Link)
	return getRandom(linksTable, l.GID).
		Bind(l)
}

func InsertLink(gid, url, text string) *sqlf.Stmt {
	return sqlf.
		InsertInto(linksTable).
		Set("gid", gid).
		Set("url", url).
		Set("text", text).
		Clause("ON CONFLICT url,gid DO UPDATE SET text = links.text")
}

func DeleteLink(gid, url string) *sqlf.Stmt {
	return sqlf.
		DeleteFrom(linksTable).
		Where("url = ? AND gid = ?", url, gid)
}

func SearchLink(gid, term string) *sqlf.Stmt {
	return sqlf.
		Select("url", "text").
		From(linksTable).
		Where("text LIKE ? AND gid = ?", "%"+term+"%", gid)
}
