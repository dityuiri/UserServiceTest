// This file contains the repository implementation layer.
package repository

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Repository struct {
	Db *sql.DB
}

type NewRepositoryOptions struct {
	Dsn string
}

func NewRepository(opts NewRepositoryOptions) *Repository {
	db, err := sql.Open("postgres", opts.Dsn)
	if err != nil {
		panic(err)
	}
	return &Repository{
		Db: db,
	}
}
