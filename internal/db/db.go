package db

import (
	"context"
	"database/sql"

	"github.com/leporo/sqlf"
)

type TableTypes interface {
	Abraxoides | Links | Media | Callout | Groups
}

func Query(ctx context.Context, db *sql.DB, stmt *sqlf.Stmt) error {
	return stmt.
		QueryRowAndClose(ctx, db)
}

func QueryOne[T TableTypes](ctx context.Context, db *sql.DB, stmt *sqlf.Stmt, t *T) error {
	return stmt.
		Bind(t).
		QueryRow(ctx, db)
}

func QueryMany[T TableTypes](ctx context.Context, db *sql.DB, stmt sqlf.Stmt, res []*T) (err error) {
	record := new(T)

	err = stmt.Bind(record).QueryAndClose(ctx, db, func(rows *sql.Rows) {
		for rows.Next() {
			if rows.Err() != nil {
				continue
			}

			res = append(res, record)
		}
	})

	return err
}

func QueryString(ctx context.Context, db *sql.DB, stmt *sqlf.Stmt) (string, error) {
	var r string
	res, err := QueryStrings(ctx, db, stmt)
	if err != nil {
		return r, err
	}

	if len(res) > 0 {
		return res[0], err
	}

	return r, err
}

func QueryStrings(ctx context.Context, db *sql.DB, stmt *sqlf.Stmt) (results []string, err error) {
	stmtErr := stmt.QueryAndClose(ctx, db, func(rows *sql.Rows) {
		for rows.Next() {
			var result string
			if e := rows.Scan(&result); e != nil {
				err = e
				continue
			}
			results = append(results, result)
		}

		if e := rows.Err(); e != nil {
			err = e
		}
	})
	if stmtErr != nil {
		return results, stmtErr
	}

	return results, err
}
