package db

import (
	"fmt"

	"github.com/leporo/sqlf"
)

var (
	mediaTable       = "media"
	mediaCreateTable = `
CREATE TABLE IF NOT EXISTS media
(	data        VARCHAR(512) NOT NULL
,	description TEXT
,	kind        VARCHAR(64)  NOT NULL
,	gid         VARCHAR(64)  NOT NULL
,	PRIMARY KEY(data,gid)
,	FOREIGN KEY(gid) REFERENCES groups
);`
)

type Media struct {
	Kind        string
	Description string
	Data        string
	GID         string
}

func (m *Media) Random() *sqlf.Stmt {
	return getRandom(mediaTable, m.GID).
		Select("kind", m.Kind).
		Select("description", m.Description).
		Select("data", m.Data)
}

func (m *Media) Delete() *sqlf.Stmt {
	return sqlf.DeleteFrom(mediaTable).
		Where("data = ? AND gid = ?", m.Data, m.GID)
}

func (m *Media) Insert() *sqlf.Stmt {
	return sqlf.
		InsertInto(mediaTable).
		Set("kind", m.Kind).
		Set("description", m.Description).
		Set("data", m.Data).
		Set("gid", m.GID).
		Clause(fmt.Sprintf("ON CONFLICT data,gid DO UPDATE SET description = %s.description", linksTable))
}

func (m *Media) Search(term string) *sqlf.Stmt {
	likeTerm := "%" + term + "%"
	return sqlf.
		Select("data, description, kind, gid").
		Where("description LIKE ? AND ", likeTerm)
}
