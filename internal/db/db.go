package db

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Client struct {
	db    *sqlx.DB
	stmts map[string]*sqlx.Stmt
	mu    sync.RWMutex
}

func (c *Client) DB() *sqlx.DB {
	return c.db
}

func (c *Client) prepareStmt(stmt string) (*sqlx.Stmt, error) {
	c.mu.RLock()
	prepared, ok := c.stmts[stmt]
	c.mu.RUnlock()
	if ok {
		return prepared, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if prepared, ok := c.stmts[stmt]; ok {
		return prepared, nil
	}

	s, err := c.db.Preparex(stmt)
	if err != nil {
		return nil, err
	}

	c.stmts[stmt] = s
	return s, nil
}

var client *Client

func Default() *Client {
	return client
}

func Open(filename string) error {
	var err error
	uri := fmt.Sprintf("file:%s?cache=private&mode=rw&_txlock=immediate&_journal_mode=WAL", filename)
	sqlDB, err := sqlx.Connect("sqlite3", uri)
	if err != nil {
		return err
	}
	client = &Client{
		db:    sqlDB,
		stmts: make(map[string]*sqlx.Stmt),
	}
	return client.db.Ping()
}

func Close() error {
	if client == nil {
		return nil
	}
	client.mu.Lock()
	client.stmts = make(map[string]*sqlx.Stmt)
	client.mu.Unlock()
	err := client.db.Close()
	client = nil
	return err
}
