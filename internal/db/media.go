package db

import (
	"context"
	"fmt"
	"time"

	"github.com/moolite/bot/internal/utils"
)

var (
	mediaTable = "media"
)

type Media struct {
	RowID       int64     // NOTE: needed by callback queries
	GID         int64     `db:"gid"`
	Data        string    `db:"data"`
	Kind        string    `db:"kind"`
	Description string    `db:"description"`
	Score       int       `db:"score"`
	CreatedAt   time.Time `db:"created_at"`
}

type MediaFts struct {
	RowID       int64
	Description string `db:"description"`
	GID         int64  `db:"gid"`
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

	return q.GetContext(ctx, m, m.Data, m.GID)
}

func SelectOneMediaByRowID(ctx context.Context, m *Media) error {
	q, err := prepareStmt(
		`SELECT rowid,data,description,gid,kind,score FROM ` + mediaTable + `
		WHERE rowid=? LIMIT 1`,
	)
	if err != nil {
		return err
	}

	return q.GetContext(ctx, m, m.RowID)
}

func SelectAllMediaGroup(ctx context.Context, gid string) ([]Media, error) {
	var results []Media
	q, err := prepareStmt(
		`SELECT rowid,data,description,gid,kind,score FROM ` + mediaTable + ` WHERE gid=?`,
	)
	if err != nil {
		return results, err
	}

	return results, q.SelectContext(ctx, &results, gid)
}

func SelectAllMedia(ctx context.Context) ([]Media, error) {
	results := []Media{}
	q, err := prepareStmt(
		`SELECT rowid,data,description,gid,kind,score FROM ` + mediaTable,
	)
	if err != nil {
		return results, err
	}

	return results, q.Select(&results)
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

	return q.GetContext(ctx, m, m.GID, m.Kind, m.GID, m.Kind)
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

	return q.GetContext(ctx, m, m.GID, m.GID)
}

func SelectMediaTop(ctx context.Context, gid int64, top int) ([]Media, error) {
	q, err := prepareStmt(
		`SELECT rowid,gid,data,description,kind,score FROM ` + mediaTable + `
		 WHERE gid=?
		 AND score > 0
		 ORDER BY score DESC
		 LIMIT ?`,
	)
	if err != nil {
		return nil, err
	}

	res := []Media{}
	return res, q.SelectContext(ctx, &res, gid, top)
}

func SelectMediaBottom(ctx context.Context, gid int64, top int) ([]Media, error) {
	q, err := prepareStmt(
		`SELECT rowid,gid,data,description,kind,score FROM ` + mediaTable + `
		 WHERE gid=?
		 AND score > 0
		 ORDER BY score ASC
		 LIMIT ?`,
	)
	if err != nil {
		return nil, err
	}

	res := []Media{}
	return res, q.SelectContext(ctx, &res, gid, top)
}

func SearchRandomMedia(ctx context.Context, m *Media, term string) error {
	q, err := prepareStmt(
		`SELECT media.rowid,media.gid,media.data,media.kind,media.description,media.score
		 FROM media_fts
		 JOIN media ON media.rowid = media_fts.rowid
		 WHERE media_fts.description MATCH ?
		   AND media_fts.gid=?
		LIMIT 1
		OFFSET ABS(RANDOM()
			% MAX((SELECT COUNT(*) FROM media WHERE gid=?),1))`,
	)
	if err != nil {
		return err
	}

	term = utils.CleanText(term)
	stringGid := fmt.Sprint(m.GID) // FTS tables are text by default
	return q.GetContext(ctx, m, term, stringGid, stringGid)
}

func SearchMedia(ctx context.Context, gid int64, term string, offset int) ([]Media, error) {
	results := []Media{}

	q, err := prepareStmt(
		`SELECT media.rowid,media.gid,media.data,media.kind,media.description,media.score
		 FROM media_fts
		 JOIN media ON media.rowid = media_fts.rowid
		 WHERE media_fts.description MATCH ?
		   AND media_fts.gid=?
		ORDER BY media.score DESC
		LIMIT 6
		OFFSET ?`,
	)
	if err != nil {
		return results, err
	}

	term = utils.CleanText(term)
	stringGid := fmt.Sprint(gid) // FTS tables are text by default
	return results, q.Select(&results, term, stringGid, offset)
}

func DeleteMedia(ctx context.Context, m *Media) error {
	q, err := prepareStmt(
		`DELETE FROM ` + mediaTable + ` WHERE gid=? AND data=?`,
	)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, m.GID, m.Data)
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
