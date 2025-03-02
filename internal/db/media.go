package db

import (
	"context"
	"fmt"
)

var (
	mediaTable = "media"
)

type Media struct {
	RowID       int64  // NOTE: needed only by callback queries
	GID         int64  `db:"gid"`
	Data        string `db:"data"`
	Kind        string `db:"kind"`
	Description string `db:"description"`
	Score       int    `db:"score"`
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
		`INSERT OR REPLACE INTO ` + mediaTable + ` (gid,data,kind,description,score)
		VALUES(?,?,?,?,?)`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, m.GID, m.Data, m.Kind, m.Description, m.Score)
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

func UpdateMediaScoreByRowID(ctx context.Context, m *Media) error {
	q, err := prepareStmt(
		`UPDATE ` + mediaTable + ` SET score=? WHERE rowid=?`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, m.Score, m.RowID)
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
		`SELECT rowid,data,description,gid,kind,score FROM ` + mediaTable + `
		WHERE data=? AND gid=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	row := q.QueryRow(m.Data, m.GID)
	return row.Scan(&m.RowID, &m.Data, &m.Description, &m.GID, &m.Kind, &m.Score)
}

func SelectOneMediaByRowID(ctx context.Context, m *Media) error {
	q, err := prepareStmt(
		`SELECT rowid,data,description,gid,kind,score FROM ` + mediaTable + `
		WHERE rowid=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	row := q.QueryRow(m.RowID)
	return row.Scan(&m.RowID, &m.Data, &m.Description, &m.GID, &m.Kind, &m.Score)
}

func SelectAllMediaGroup(ctx context.Context, gid string) ([]*Media, error) {
	var results []*Media
	q, err := prepareStmt(
		`SELECT rowid,data,description,gid,kind,score FROM ` + mediaTable + ` WHERE gid=?`,
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
		m := new(Media)

		if err = rows.Scan(&m.RowID, &m.Data, &m.Description, &m.GID, &m.Kind, &m.Score); err != nil {
			return results, err
		}
		results = append(results, m)
	}

	return results, nil
}

func SelectAllMedia(ctx context.Context) ([]*Media, error) {
	var results []*Media
	q, err := prepareStmt(
		`SELECT rowid,data,description,gid,kind,score FROM ` + mediaTable,
	)
	if err != nil {
		return results, err
	}

	rows, err := q.Query()
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		m := new(Media)

		if err = rows.Scan(&m.RowID, &m.Data, &m.Description, &m.GID, &m.Kind, &m.Score); err != nil {
			return results, err
		}
		results = append(results, m)
	}

	return results, nil
}

func SelectRandomMediaKind(ctx context.Context, m *Media) error {
	q, err := prepareStmt(
		`SELECT rowid,gid,data,description,kind,score FROM media
		 WHERE gid=? AND kind=?
		 LIMIT 1
		 OFFSET ABS(RANDOM()
			% MAX((SELECT COUNT(*) FROM media WHERE gid=? AND kind=?),1))`,
	)
	if err != nil {
		return err
	}

	row := q.QueryRowContext(ctx, m.GID, m.Kind, m.GID, m.Kind)
	return row.Scan(&m.RowID, &m.GID, &m.Data, &m.Description, &m.Kind, &m.Score)
}

func SelectRandomMedia(ctx context.Context, m *Media) error {
	if len(m.Kind) > 0 {
		return SelectRandomMediaKind(ctx, m)
	}

	q, err := prepareStmt(
		`SELECT rowid,gid,data,description,kind,score FROM media
		 WHERE gid=?
		 LIMIT 1
		 OFFSET ABS(RANDOM()
			% MAX((SELECT COUNT(*) FROM media WHERE gid=?),1))`,
	)
	if err != nil {
		return err
	}

	row := q.QueryRowContext(ctx, m.GID, m.GID)
	return row.Scan(&m.RowID, &m.GID, &m.Data, &m.Description, &m.Kind, &m.Score)
}

func SearchMedia(ctx context.Context, gid, term string) (*Media, error) {
	likeTerm := fmt.Sprintf("%%%s%%", term)
	m := new(Media)

	q, err := prepareStmt(
		`SELECT rowid,gid,kind,data,description,score FROM ` + mediaTable + `
		WHERE description LIKE ? AND gid=?`,
	)
	if err != nil {
		return m, err
	}

	row := q.QueryRowContext(ctx, likeTerm, m.GID)
	if err := row.Scan(&m.RowID, &m.GID, &m.Data, &m.Description, &m.Kind, &m.Score); err != nil {
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
