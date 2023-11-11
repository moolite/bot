package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/leporo/sqlf"
	"github.com/rs/zerolog/log"
)

var (
	mediaTable       = "media"
	mediaCreateTable = `
CREATE TABLE IF NOT EXISTS media
( data        VARCHAR(512) NOT NULL
, kind        VARCHAR(64)  NOT NULL
, gid         VARCHAR(64)  NOT NULL
, description TEXT
, PRIMARY KEY(data,gid)
, FOREIGN KEY(gid) REFERENCES groups(gid)
);`
)

type Media struct {
	GID         string `db:"gid"`
	Data        string `db:"data"`
	Kind        string `db:"kind"`
	Description string `db:"description"`
}

func (m *Media) Clone() *Media {
	return &Media{
		GID:         m.GID,
		Data:        m.Data,
		Kind:        m.Kind,
		Description: m.Description,
	}
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

func SelectOneMediaByData(ctx context.Context, m *Media) error {
	q := sqlf.
		From(mediaTable).
		Select("data").To(&m.Data).
		Select("description").To(&m.Description).
		Select("gid").To(&m.GID).
		Select("kind").To(&m.Kind).
		Where("data = ? AND gid = ?", m.Data, m.GID).
		Limit(1)

	log.Debug().
		Str("stmt", q.String()).
		Interface("args", q.Args()).
		Msg("statement")

	return q.QueryRowAndClose(ctx, dbc)
}

func SelectAllMedia(ctx context.Context, gid string) ([]*Media, error) {
	var results []*Media
	m := &Media{}
	q := sqlf.
		From(mediaTable).
		Select("data").To(&m.Data).
		Select("description").To(&m.Description).
		Select("gid").To(&m.GID).
		Select("kind").To(&m.Kind).
		Where("gid = ?", gid)
	err := q.QueryAndClose(ctx, dbc, func(rows *sql.Rows) {
		results = append(results, m.Clone())
	})
	return results, err
}

func SelectRandomMedia(ctx context.Context, m *Media) error {
	q := getRandom(mediaTable, m.GID).
		Select("data").To(&m.Data).
		Select("description").To(&m.Description).
		Select("gid").To(&m.GID).
		Select("kind").To(&m.Kind).
		Where("kind = ?", m.Kind)

	log.Debug().
		Str("stmt", q.String()).
		Interface("args", q.Args()).
		Msg("SelectRandomMedia")

	return q.QueryRowAndClose(ctx, dbc)
}

func SearchMedia(ctx context.Context, gid, term string) (*Media, error) {
	likeTerm := fmt.Sprintf("%%%s%%", term)

	m := new(Media)

	q := sqlf.
		From(mediaTable).
		Select("gid").To(&m.GID).
		Select("kind").To(&m.Kind).
		Select("data").To(&m.Data).
		Select("description").To(&m.Description).
		Where("description LIKE ? AND gid = ?", likeTerm, gid)

	if err := q.QueryRowAndClose(ctx, dbc); err != nil {
		return nil, err
	}
	return m, nil
}

func DeleteMedia(ctx context.Context, media *Media) error {
	q := sqlf.
		DeleteFrom(mediaTable).
		Where("data = ? AND gid = ?", media.Data, media.GID).
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
