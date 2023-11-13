package db

import (
	"context"
	"fmt"
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
	q, err := prepr(`SELECT url,text,gid FROM ` + linksTable + ` WHERE gid=? LIMIT 1`)
	if err != nil {
		return err
	}

	row := q.QueryRowContext(ctx, l.GID)

	return row.Scan(&l.URL, &l.Text, &l.GID)
}

func SearchLinks(ctx context.Context, gid, term string) (links []*Link, err error) {
	likeTerm := fmt.Sprintf("%%%s%%", term)
	q, err := prepr(`SELECT text,url,gid FROM ` + linksTable + ` WHERE text LIKE ? AND gid=?`)
	if err != nil {
		return links, err
	}

	rows, err := q.QueryContext(ctx, likeTerm, gid)
	if err != nil {
		return links, err
	}
	defer rows.Close()

	for rows.Next() {
		var l *Link
		if err := rows.Scan(&l.Text, &l.URL, &l.GID); err != nil {
			return links, err
		} else {
			links = append(links, l)
		}
	}

	return links, nil
}

func InsertLink(ctx context.Context, l *Link) error {
	q, err := prepr(`INSERT INTO ` + linksTable + ` (url,text,gid) VALUES(?,?,?)
	  ON CONFLICT(url,gid) DO UPDATE SET text=` + linksTable + `.text`)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, l.URL, l.Text, l.GID)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		return ErrInsert
	}
	return nil
}

func DeleteLink(ctx context.Context, l *Link) error {
	q, err := prepr(`DELETE FROM ` + linksTable + ` WHERE url=? AND gid=?`)
	if err != nil {
		return err
	}

	r, err := q.ExecContext(ctx, l.URL, l.GID)
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
