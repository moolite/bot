package db

import (
	"errors"
)

var (
	ErrNotFound error = errors.New("record not found")
	ErrInsert   error = errors.New("record insertion failure")
	ErrDelete   error = errors.New("error deleting record")
)
