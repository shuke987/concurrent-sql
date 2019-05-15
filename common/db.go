package common

import "database/sql"

var gDB *sql.DB

func InitDB() error {
	if db, err := sql.Open("mysql", "root@/test"); err != nil {
		return err
	} else {
		gDB = db
		return nil
	}
}

func ReleaseDB() error {
	if err := gDB.Close(); err != nil {
		return err
	}
	gDB = nil
	return nil
}

func DB() *sql.DB {
	return gDB
}
