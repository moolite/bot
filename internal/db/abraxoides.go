package db

import (
	"context"
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
	q, err := prepr(`SELECT gid,abraxas,kind FROM ` + abraxoidesTable + ` WHERE gid=? AND abraxas=? LIMIT 1`)
	if err != nil {
		return err
	}

	row := q.QueryRowContext(ctx, a.GID, a.Abraxas)

	return row.Scan(&a.GID, &a.Abraxas, &a.Kind)
}

func SelectAbraxoides(ctx context.Context, gid string) ([]*Abraxas, error) {
	var abraxoides []*Abraxas

	q, err := prepr(`SELECT abraxas,kind,gid FROM ` + abraxoidesTable + ` WHERE gid=?`)
	if err != nil {
		return abraxoides, err
	}

	rows, err := q.QueryContext(ctx, gid)
	if err != nil {
		return abraxoides, err
	}

	for rows.Next() {
		var a *Abraxas
		err := rows.Scan(&a.Abraxas, &a.Kind, &a.GID)
		if err != nil {
			return abraxoides, err
		}
		abraxoides = append(abraxoides, a)
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
	q, err := prepr(`INSERT INTO ` + abraxoidesTable + ` (gid,abraxas,kind) VALUES (?,?,?)
	ON CONFLICT(abraxas,gid) DO UPDATE SET kind=` + abraxoidesTable + `.kind`)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, a.GID, a.Abraxas, a.Kind)
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
