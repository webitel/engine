package store

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type Driver interface {
	Begin(opts *sql.TxOptions) (*sqlx.Tx, error)
	Exec(query string, args interface{}) (sql.Result, error)
	SelectOne(holder interface{}, query string, args interface{}) error
	Select(dest interface{}, query string, args interface{}) ([]interface{}, error)
	SelectInt(query string, args interface{}) (int64, error)
	SelectNullInt(query string, args interface{}) (sql.NullInt64, error)
	TraceOn()
}
