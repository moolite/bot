package db

import (
	"context"
)

var (
	abraxoidesTable string = "abraxoides"
)

type Abraxas struct {
	GID     int64  `db:"gid"`
	Abraxas string `db:"abraxas"`
	Kind    string `db:"kind"`
}

func (a *Abraxas) Clone() *Abraxas {
	return &Abraxas{
		Abraxas: a.Abraxas,
		Kind:    a.Kind,
		GID:     a.GID,
	}
}

func SelectOneAbraxasByAbraxas(ctx context.Context, a *Abraxas) error {
	q, err := prepareStmt(
		`SELECT gid,abraxas,kind FROM ` + abraxoidesTable + `
		WHERE gid=? AND abraxas LIKE ? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	return q.GetContext(ctx, a, a.GID, a.Abraxas)
}

func SelectAbraxoides(ctx context.Context, gid string) ([]Abraxas, error) {
	abraxoides := []Abraxas{}

	q, err := prepareStmt(
		`SELECT abraxas,kind,gid FROM ` + abraxoidesTable + ` WHERE gid=?`,
	)
	if err != nil {
		return abraxoides, err
	}

	return abraxoides, q.SelectContext(ctx, &abraxoides, gid)
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
	q, err := prepareStmt(
		`INSERT OR REPLACE INTO ` + abraxoidesTable + ` (gid,abraxas,kind) VALUES (?,?,?)`,
	)
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

func DeleteAbraxas(ctx context.Context, a *Abraxas) error {
	q, err := prepareStmt(
		`DELETE FROM ` + abraxoidesTable + ` WHERE gid=? AND abraxas=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, a.GID, a.Abraxas)
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
