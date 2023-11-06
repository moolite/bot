package db

import (
	"context"
	"fmt"

	"github.com/leporo/sqlf"
)

var (
	mediaTable       = "media"
	mediaCreateTable = `
CREATE TABLE IF NOT EXISTS media
( data        VARCHAR(512) NOT NULL
, description TEXT
, kind        VARCHAR(64)  NOT NULL
, gid         VARCHAR(64)  NOT NULL
, PRIMARY KEY(data,gid)
, FOREIGN KEY(gid) REFERENCES groups
);`
)

type Media struct {
	Kind        string `db:"kind"`
	Description string `db:"description"`
	Data        string `db:"data"`
	GID         string `db:"gid"`
}

func InsertMedia(ctx context.Context, media *Media) error {
	q := sqlf.
		InsertInto(mediaTable).
		Set("gid", media.GID).
		Set("data", media.Data).
		Set("kind", media.Kind).
		Set("description", media.Description).
		Clause(
			"ON CONFLICT(data,gid) DO UPDATE SET description = media.description, kind = media.kind")

	res, err := q.Exec(ctx, dbc)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil
	}
	if n == 0 {
		return ErrInsert
	}

	return nil
}

func SelectRandomMedia(ctx context.Context, m *Media) error {
	q := getRandom(mediaTable, m.GID).
		Select("kind", m.Kind).
		Select("description", m.Description).
		Select("data", m.Data).
		Where("kind = ?", m.Kind)

	return q.QueryRow(ctx, dbc)
}

func SearchMedia(ctx context.Context, gid, term string) (*Media, error) {
	likeTerm := fmt.Sprintf("%%%s%%", term)

	m := new(Media)

	q := sqlf.
		Select("kind", m.Kind).
		Select("description", m.Description).
		Select("data", m.Data).
		From(mediaTable).
		Where("description LIKE ? AND gid = ?", likeTerm, gid)

	if err := q.QueryRow(ctx, dbc); err != nil {
		return nil, err
	}
	return m, nil
}

func DeleteMedia(ctx context.Context, media *Media) error {
	q := sqlf.
		DeleteFrom(mediaTable).
		Where("data = ? AND gid = ?", media.Data, media.GID).
		Limit(1)

	res, err := q.Exec(ctx, dbc)
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
