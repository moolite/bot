package db

import (
	"context"
	"database/sql"

	"github.com/leporo/sqlf"
)

var (
	abraxoidesTable       string = "abraxoides"
	abraxoidesCreateTable string = `
CREATE TABLE IF NOT EXISTS abraxoides
( abraxas VARCHAR(128) NOT NULL
, kind    VARCHAR(64)  NOT NULL
, gid     VARCHAR(64)  NOT NULL
, PRIMARY KEY(abraxas,gid)
, FOREIGN KEY(gid) REFERENCES groups
);
`
)

type Abraxas struct {
	Abraxas string `db:"abraxas"`
	Kind    string `db:"kind"`
	GID     string `db:"gid"`
}

func (a *Abraxas) Clone() *Abraxas {
	return &Abraxas{
		Abraxas: a.Abraxas,
		Kind:    a.Kind,
		GID:     a.GID,
	}
}

func SelectOneAbraxas(ctx context.Context, a *Abraxas) error {
	q := sqlf.
		From(abraxoidesTable).
		Select("abraxas").To(&a.Abraxas).
		Select("kind").To(&a.Kind).
		Where("gid = ?", a.GID).
		Where("abraxas = ?", a.Abraxas).
		Limit(1)

	return q.QueryRow(ctx, dbc)
}

func SelectAbraxoides(ctx context.Context, gid string) ([]string, error) {
	var abraxas string
	var abraxoides []string

	q := sqlf.
		Select("abraxas").To(&abraxas).
		Where("gid = ?", gid)

	err := q.Query(ctx, dbc, func(rows *sql.Rows) {
		abraxoides = append(abraxoides, abraxas)
	})
	if err != nil {
		return nil, err
	}

	return abraxoides, nil
}

func InsertAbraxas(ctx context.Context, gid, abraxas, kind string) error {
	q := sqlf.
		InsertInto(abraxoidesTable).
		Set("gid", gid).
		Set("abraxas", abraxas).
		Set("kind", kind).
		Clause(
			"ON CONFLICT(abraxas,gid) DO UPDATE SET kind = abraxoides.kind")

	if res, err := q.Exec(ctx, dbc); err != nil {
		return err
	} else if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrInsert
	}

	return nil
}
