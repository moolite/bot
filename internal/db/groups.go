package db

import "github.com/leporo/sqlf"

type Groups struct {
	GID   string `db:"gid"`
	Title string `db:"title"`
}

var (
	groupsTable       string = "groups"
	groupsCreateTable string = `
CREATE TABLE IF NOT EXISTS groups
(	gid VARCHAR(64) NOT NULL
,	title text
,	PRIMARY KEY(gid)
);`
)

func (g *Groups) One(gid string) *sqlf.Stmt {
	return sqlf.
		Select("gid, title").
		From(groupsTable).
		Where("gid = ?", gid)
}

func (g *Groups) Insert(gid, title string) *sqlf.Stmt {
	return sqlf.
		InsertInto(groupsTable).
		Set("gid", gid).
		Set("title", title)
}

func (g *Groups) All() *sqlf.Stmt {
	return sqlf.
		Select("gid, title").
		From(groupsTable)
}
