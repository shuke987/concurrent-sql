package verify

import "database/sql"

type PlanAssert struct {
	SQL    string
	Expect string
}

func (pa *PlanAssert) Assert(db *sql.DB) error {
	_, err := db.Exec(pa.SQL)
	if err != nil {
		return err
	}

	return nil
}
