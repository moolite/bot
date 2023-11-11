package db

import (
	"github.com/leporo/sqlf"
)

func getRandom(table, gid string) *sqlf.Stmt {
	return sqlf.
		From(table).
		Where("gid = ?", gid).
		Clause(
			"LIMIT 1 OFFSET ABS(RANDOM()) % MAX((SELECT COUNT(*) FROM " + table + " WHERE gid='" + gid + "'),1)",
		)
}
