package db

import (
	"github.com/leporo/sqlf"
)

var (
	abraxoidexTable       string = "abraxoides"
	abraxoidesCreateTable string = `
CREATE OR UPDATE TABLE abraxoides
(
	abraxas VARCHAR 128 NOT NULL,
	kind    VARCHAR 64  NOT NULL,
	gid     VARCHAR 64  NOT NULL,

	PRIMARY KEY(abraxas,gid),
	FOREIGN KEY(gid) REFERENCES groups,
)`
)

type Abraxoides struct {
	Abraxas string `db:"abraxas"`
	Kind    string `db:"kind"`
	GID     string `db:"gid"`
}

func (a *Abraxoides) Clone() *Abraxoides {
	return &Abraxoides{
		Abraxas: a.Abraxas,
		Kind:    a.Kind,
		GID:     a.GID,
	}
}

func (a *Abraxoides) One() *sqlf.Stmt {
	return sqlf.
		Select("abraxas", a.Abraxas).
		Select("kind", a.Kind).
		Select("gid", a.GID).
		From("abraxoides").
		Where("abraxas = ?", a.Abraxas)
}

func (a *Abraxoides) Insert() *sqlf.Stmt {
	return sqlf.
		InsertInto("abraxoides").
		Set("gid", a.GID).
		Set("kind", a.Kind).
		Set("abraxas", a.Abraxas).
		Clause("ON CONFLICT abraxas,gid DO UPDATE SET kind = abraxoides.kind")
}

func (a *Abraxoides) AllKeywords() *sqlf.Stmt {
	return sqlf.
		Select("abraxas").
		From("abraxoides").
		Where("gid = ?", a.GID)
}
