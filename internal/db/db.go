package db

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
)

var dbc *sql.DB
var stmts map[string]*sql.Stmt = make(map[string]*sql.Stmt)

func Open(filename string) error {
	var err error
	uri := fmt.Sprintf("%s?cache=shared", filename)
	dbc, err = sql.Open("sqlite3", uri)
	if err != nil {
		return err
	}
	return dbc.Ping()
}

func Close() error {
	// reset prepared statements
	stmts = make(map[string]*sql.Stmt)

	return dbc.Close()
}

func onRow(rows *sql.Rows) {
	columns, err := rows.Columns()
	if err != nil {
		log.Debug().Err(err).Msg("error fetching columns")
	} else {
		log.Debug().Strs("columns", columns).Msg("rows")
	}
}

func prepr(stmt string) (*sql.Stmt, error) {
	if prepared, ok := stmts[stmt]; ok {
		return prepared, nil
	}

	s, err := dbc.Prepare(stmt)
	if err != nil {
		return nil, err
	}

	stmts[stmt] = s
	return s, nil
}
