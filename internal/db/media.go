package db

import (
	"context"
	"fmt"
)

var (
	mediaTable = "media"
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

func InsertMedia(ctx context.Context, m *Media) error {
	q, err := prepareStmt(
		`INSERT OR REPLACE INTO `+mediaTable+` (gid,data,kind,description)
		VALUES(?,?,?,?)`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, m.GID, m.Data, m.Kind, m.Description)
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
	q, err := prepareStmt(
		`SELECT data,description,gid,kind FROM ` + mediaTable + `
		WHERE data=? AND gid=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	row := q.QueryRow(m.Data, m.GID)
	return row.Scan(&m.Data, &m.Description, &m.GID, &m.Kind)
}

func SelectAllMedia(ctx context.Context, gid string) ([]Media, error) {
	var results []Media
	q, err := prepareStmt(
		`SELECT data,description,gid,kind FROM ` + mediaTable + ` WHERE gid=?`
	)
	if err != nil {
		return results, err
	}

	rows, err := q.Query(gid)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var m *Media
		err = rows.Scan(&m.Data, &m.Description, &m.GID, &m.Kind)
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

func SelectRandomMedia(ctx context.Context, m *Media) error {
	q, err := prepareStmt(
		`SELECT gid,data,description,kind FROM media
		WHERE gid=? AND kind=?
		LIMIT 1
		OFFSET ABS(RANDOM()
			% MAX((SELECT COUNT(*) FROM media WHERE gid=? AND kind=?),1))`,
	)
	if err != nil {
		return err
	}

	row := q.QueryRowContext(ctx, m.GID, m.Kind, m.GID, m.Kind)
	if err := row.Scan(&m.GID, &m.Data, &m.Description, &m.Kind); err != nil {
		return err
	}
	return nil
}

func SearchMedia(ctx context.Context, gid, term string) (*Media, error) {
	likeTerm := fmt.Sprintf("%%%s%%", term)
	var m *Media

	q, err := prepareStmt(
		`SELECT gid,kind,data,description FROM ` + mediaTable + `
		WHERE description LIKE ? AND gid=?`,
	)
	if err != nil {
		return m, err
	}

	row := q.QueryRowContext(ctx, likeTerm, m.GID)
	if err := row.Scan(&m.GID, &m.Data, &m.Description, &m.Kind); err != nil {
		return m, err
	}
	return m, nil
}

func DeleteMedia(ctx context.Context, m *Media) error {
	q, err := prepareStmt(
		`DELETE FROM ` + mediaTable + ` WHERE data=? AND gid=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, m.Data, m.GID)
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
