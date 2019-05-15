package verify

import (
	"database/sql"
	"log"
)

type AssertNoError struct {
	SQL string
}

func (check *AssertNoError) Assert(db *sql.DB) (err error) {
	if _, err = db.Exec(check.SQL); err != nil {
		log.Printf("assert no error failed, %s, %s", check.SQL, err)
	}

	return
}
