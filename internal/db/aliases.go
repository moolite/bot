package db

import "context"

const aliasTable string = `aliases`

type Alias struct {
	Name   string `db:"name"`
	Target string `db:"target"`
}

func SelectAlias(ctx context.Context, alias *Alias) error {
	q, err := prepareStmt(
		`SELECT name,target FROM ` + aliasTable + ` WHERE name=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	return q.GetContext(ctx, alias.Name)
}

func SelectAllAliases(ctx context.Context) ([]Alias, error) {
	ret := []Alias{}

	q, err := prepareStmt(
		`SELECT name,target FROM ` + aliasTable,
	)
	if err != nil {
		return ret, err
	}

	return ret, q.SelectContext(ctx, &ret)
}

func InsertAlias(ctx context.Context, alias *Alias) error {
	q, err := prepareStmt(
		`INSERT OR INTO ` + aliasTable + `(name,target) VALUES(?,?) ON CONFLICT(name) DO UPDATE SET name=?`,
	)
	if err != nil {
		return err
	}

	if res, err := q.ExecContext(ctx, alias.Name, alias.Target, alias.Name); err != nil {
		return err
	} else if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrInsert
	}
	return nil
}
