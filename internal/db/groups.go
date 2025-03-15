package db

import (
	"context"
)

var (
	groupsTable string = "groups"
)

type Group struct {
	GID   int64  `db:"gid"`
	Title string `db:"title"`
}

func (g *Group) Clone() *Group {
	return &Group{
		GID:   g.GID,
		Title: g.Title,
	}
}

func SelectOneGroup(ctx context.Context, g *Group) error {
	q, err := prepareStmt(
		`SELECT title FROM ` + groupsTable + ` WHERE gid=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	row := q.QueryRowContext(ctx, g.GID)

	if err := row.Scan(&g.Title); err != nil {
		return err
	}
	return nil
}

func SelectAllGroups(ctx context.Context) ([]*Group, error) {
	var ret []*Group

	q, err := prepareStmt(
		`SELECT gid,title FROM ` + groupsTable,
	)
	if err != nil {
		return ret, err
	}

	rows, err := q.QueryContext(ctx)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		g := &Group{}
		if err := rows.Scan(&g.GID, &g.Title); err != nil {
			return ret, err
		} else {
			ret = append(ret, g)
		}
	}
	return ret, nil
}

func InsertGroup(ctx context.Context, gid int64, title string) error {
	q, err := prepareStmt(
		`INSERT INTO ` + groupsTable + ` (gid,title) VALUES(?,?)
		ON CONFLICT(gid) DO UPDATE SET title=excluded.title`,
	)
	if err != nil {
		return err
	}

	if res, err := q.ExecContext(ctx, gid, title); err != nil {
		return err
	} else if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrInsert
	}
	return nil
}
