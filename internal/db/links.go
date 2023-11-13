package db

import (
	"context"
	"database/sql"

	"github.com/leporo/sqlf"
	"github.com/rs/zerolog/log"
)

var (
	linksTable       string = "links"
	linksCreateTable string = `
CREATE TABLE IF NOT EXISTS links
( url VARCHAR(256) NOT NULL
, text TEXT
, gid VARCHAR(64) NOT NULL
, PRIMARY KEY(url,gid)
, FOREIGN KEY(gid) REFERENCES groups
);
`
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

func SelectLinkByURL(ctx context.Context, l *Link) error {
	q := sqlf.
		From(linksTable).
		Select("url").To(&l.URL).
		Select("gid").To(&l.GID).
		Select("text").To(&l.Text).
		Where("gid = ?", l.GID).
		Limit(1)

	return q.QueryRowAndClose(ctx, dbc)
}

func SelectRandomLink(ctx context.Context, l *Link) error {
	q := getRandom(linksTable, l.GID).
		Select("url", l.URL).
		Select("text", l.Text)

	return q.QueryRowAndClose(ctx, dbc)
}

func SearchLinks(ctx context.Context, gid, term string) ([]*Link, error) {
	l := &Link{GID: gid}

	q := sqlf.
		Select("url", l.URL).
		Select("text", l.Text).
		From(linksTable).
		Where("text LIKE ? AND gid = ?", "%"+term+"%", gid)

	var results []*Link
	if err := q.QueryAndClose(ctx, dbc, func(rows *sql.Rows) {
		results = append(results, l.Clone())
	}); err != nil {
		return nil, err
	}

	return results, nil
}

func InsertLink(ctx context.Context, link *Link) error {
	q := sqlf.
		InsertInto(linksTable).
		Set("gid", link.GID).
		Set("url", link.URL).
		Set("text", link.Text).
		Clause(
			"ON CONFLICT(url,gid) DO UPDATE SET text=" + linksTable + ".text")

	log.Debug().Str("stmt", q.String()).Msg("InsertLink")

	res, err := q.ExecAndClose(ctx, dbc)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	} else if n != 1 {
		return ErrInsert
	} else {
		return nil
	}
}

func DeleteLink(ctx context.Context, l *Link) error {
	q := sqlf.
		DeleteFrom(linksTable).
		Where("url = ? AND gid = ?", l.URL, l.GID)

	log.Debug().Str("stmt", q.String()).Interface("args", q.Args()).Msg("DeleteLink")

	r, err := q.ExecAndClose(ctx, dbc)
	if err != nil {
		return err
	}
	if n, err := r.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		return ErrNotFound
	}
	return nil
}
