package db

import (
	"context"
	"database/sql"

	"github.com/leporo/sqlf"
	"github.com/rs/zerolog/log"
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

func SelectOneAbraxasByAbraxas(ctx context.Context, a *Abraxas) error {
	q := sqlf.
		From(abraxoidesTable).
		Select("gid").To(&a.GID).
		Select("abraxas").To(&a.Abraxas).
		Select("kind").To(&a.Kind).
		Where("gid = ?", a.GID).
		Where("abraxas = ?", a.Abraxas).
		Limit(1)

	log.Debug().
		Str("stmt", q.String()).
		Interface("args", q.Args()).
		Msg("SelectOneAbraxasByAbraxas statement")

	return q.QueryRowAndClose(ctx, dbc)
}

func SelectAbraxoides(ctx context.Context, gid string) ([]*Abraxas, error) {
	a := &Abraxas{}
	var abraxoides []*Abraxas

	q := sqlf.
		Select("abraxas").To(&a.Abraxas).
		Select("gid").To(&a.GID).
		Select("kind").To(&a.Kind).
		Where("gid = ?", gid)

	err := q.QueryAndClose(ctx, dbc, func(rows *sql.Rows) {
		abraxoides = append(abraxoides, a.Clone())
	})
	if err != nil {
		return nil, err
	}

	return abraxoides, nil
}

func SelectAbraxoidesAbraxas(ctx context.Context, gid string) ([]string, error) {
	var results []string
	abs, err := SelectAbraxoides(ctx, gid)
	if err != nil {
		return results, err
	}

	for _, a := range abs {
		results = append(results, a.Abraxas)
	}

	return results, err
}

func SelectAbraxoidesAbraxasKind(ctx context.Context, gid string) ([][]string, error) {
	var results [][]string
	abs, err := SelectAbraxoides(ctx, gid)
	if err != nil {
		return results, err
	}

	for _, a := range abs {
		results = append(results, []string{a.Abraxas, a.Kind})
	}

	return results, err
}

func InsertAbraxas(ctx context.Context, a *Abraxas) error {
	q := sqlf.
		InsertInto(abraxoidesTable).
		Set("gid", a.GID).
		Set("abraxas", a.Abraxas).
		Set("kind", a.Kind).
		Clause(
			"ON CONFLICT(abraxas,gid) DO UPDATE SET kind = " + abraxoidesTable + ".kind")

	log.Debug().
		Str("stmt", q.String()).
		Interface("args", q.Args()).
		Msg("InsertAbraxas")

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
