package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var dbc *sqlx.DB
var stmts map[string]*sqlx.Stmt = make(map[string]*sqlx.Stmt)
var nstmts map[string]*sqlx.NamedStmt = make(map[string]*sqlx.NamedStmt)

func Open(filename string) error {
	var err error
	uri := fmt.Sprintf("%s?cache=shared", filename)
	dbc, err = sqlx.Connect("sqlite3", uri)
	if err != nil {
		return err
	}
	return dbc.Ping()
}

func Close() error {
	// reset prepared statements
	stmts = make(map[string]*sqlx.Stmt)

	return dbc.Close()
}

func prepareStmt(stmt string) (*sqlx.Stmt, error) {
	if prepared, ok := stmts[stmt]; ok {
		return prepared, nil
	}

	s, err := dbc.Preparex(stmt)
	if err != nil {
		return nil, err
	}

	stmts[stmt] = s
	return s, nil
}

func prepareNamedStmt(stmt string) (*sqlx.NamedStmt, error) {
	if prepared, ok := nstmts[stmt]; ok {
		return prepared, nil
	}
	s, err := dbc.PrepareNamed(stmt)
	if err != nil {
		return nil, err
	}

	nstmts[stmt] = s
	return s, nil
}
