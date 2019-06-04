package ddl

import (
	"database/sql"
	"log"

	"github.com/shuke987/concurrent-sql/util"
)

type DDL struct {
	Queries []string
	DB      *sql.DB
}

func (d *DDL) Load(path string) (err error) {
	d.Queries, err = util.ReadFileLines(path)
	return err
}

func (d *DDL) Run() error {
	for _, q := range d.Queries {
		if _, err := d.DB.Exec(q); err != nil {
			log.Println("error encountered ", err)
			return err
		}
	}

	return nil
}
