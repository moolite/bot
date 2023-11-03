package db

import (
	"github.com/leporo/sqlf"
)

var (
	abraxoidexTable       string = "abraxoides"
	abraxoidesCreateTable string = `
CREATE TABLE IF NOT EXISTS abraxoides
(	abraxas VARCHAR(128) NOT NULL
,	kind    VARCHAR(64)  NOT NULL
,	gid     VARCHAR(64)  NOT NULL
,	PRIMARY KEY(abraxas,gid)
,	FOREIGN KEY(gid) REFERENCES groups
);`
)

type Abraxas struct {
	Abraxas string `db:"abraxas"`
	Kind    string `db:"kind"`
	GID     string `db:"gid"`
}

func (a *Abraxas) Clone() *Abraxas {
	return &Abraxas{
		Abraxas: a.Abraxas,
		Kind:    a.Kind,
		GID:     a.GID,
	}
}

func SelectOneAbraxas(gid, abraxas string) *sqlf.Stmt {
	return sqlf.
		Select("abraxas", "kind", "gid").
		From("abraxoides").
		Where("abraxas = ? AND gid = ?", abraxas, gid)
}

func InsertAbraxas(gid, abraxas, kind string) *sqlf.Stmt {
	return sqlf.
		InsertInto("abraxoides").
		Set("gid", gid).
		Set("abraxas", abraxas).
		Set("kind", kind).
		Clause("ON CONFLICT abraxas,gid DO UPDATE SET kind = abraxoides.kind")
}

func SelectAbraxoides(gid string) *sqlf.Stmt {
	return sqlf.
		Select("abraxas").
		From("abraxoides").
		Where("gid = ?", gid)
}
