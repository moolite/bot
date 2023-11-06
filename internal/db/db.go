package db

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
)

var dbc *sql.DB

func Connect(filename string) error {
	uri := fmt.Sprintf("%s?cache=shared", filename)
	d, err := sql.Open("sqlite3", uri)
	if err != nil {
		return err
	}

	dbc = d
	return nil
}

func onRow(rows *sql.Rows) {
	columns, err := rows.Columns()
	if err != nil {
		log.Debug().Err(err).Msg("error fetching columns")
	} else {
		log.Debug().Strs("columns", columns).Msg("rows")
	}
}
