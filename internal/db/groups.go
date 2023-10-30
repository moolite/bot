package db

import "github.com/leporo/sqlf"

type Groups struct {
	GID   string `db:"gid"`
	Title string `db:"title"`
}

var (
	groupsTable       string = "groups"
	groupsCreateTable string = `
CREATE TABLE groups
(
	gid VARCHAR 64 NOT NULL,
	title text,
	PRIMARY KEY(gid)
)
`
)

func (g *Groups) One() *sqlf.Stmt {
	return sqlf.
		Select("gid, title").
		From(groupsTable).
		Bind(g)
}

func (g *Groups) Insert() *sqlf.Stmt {
	return sqlf.
		InsertInto(groupsTable).
		Set("gid", g.GID).
		Set("title", g.Title)
}

func (g *Groups) All() *sqlf.Stmt {
	return sqlf.
		Select("gid, title").
		From(groupsTable)
}
