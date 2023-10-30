package db

import (
	"github.com/leporo/sqlf"
)

func getRandom(table, gid string) *sqlf.Stmt {
	return sqlf.
		From(table).
		Where("gid = ?", gid).
		Limit(1).
		Offset("ABS(RANDOM()) % MAX(, 1)").
		Clause(
			"OFFSET ABS(RANDOM()) % MAX((SELECT COUNT(*) FROM ? WHERE gid = ?), 1)",
			table, gid)
}
