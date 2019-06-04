package dml

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/shuke987/concurrent-sql/util"
)

type DML struct {
	SQLs    []string
	Repeats int
	DSN     string
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
			if _, err := db.Exec(q); err != nil {
				errStr := fmt.Sprintf("sql execute error: %s", err)
				log.Println(errStr)
				c <- fmt.Sprintf(errStr)
				return
			}
		}
	}
}
