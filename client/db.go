package client

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type DBClient struct {
	Connection *sql.DB
}

func (db *DBClient) Init(dsn string) (err error) {
	db.Connection, err = sql.Open("mysql", dsn)
	return err
}

func (db *DBClient) Release() {

}
