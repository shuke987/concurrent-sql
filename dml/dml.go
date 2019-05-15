package dml

import (
	"concurrent-sql/util"
	"database/sql"
	"fmt"
	"log"
)

type DML struct {
	SQLs    []string
	Repeats int
	DSN     string
	Fatal   bool
}

func (d *DML) Load(path string) (err error) {
	d.SQLs, err = util.ReadFileLines(path)
	return err
}

func (d *DML) RunAsync(c chan string, shutdown chan struct{}) {
	defer close(c)

	db, err := sql.Open("mysql", d.DSN)
	if err != nil {
		c <- fmt.Sprintf("bad database connection: %s, %s", err, d.DSN)
		return
	} else if err = db.Ping(); err != nil {
		log.Println(err)
		c <- fmt.Sprintf("ping db error, %s", err)
		_ = db.Close()
		return
	}
	defer func() {
		_ = db.Close()
	}()

	for i := 0; i < d.Repeats; i++ {
		select {
		case <-shutdown:
			return
		default:
		}

		for _, q := range d.SQLs {
			if d.Fatal {
				c <- ""
				return
			}

			if _, err := db.Exec(q); err != nil {
				c <- ""
				return
			}
		}
	}
}
