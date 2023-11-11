package db

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
)

var dbc *sql.DB

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
