package storage

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func IsErrorPGAlreadyExist(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return false
}
func IsErrorPGNotExist(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
