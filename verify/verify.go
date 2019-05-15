package verify

import (
	"database/sql"
	"log"
	"os"
)

type SQLAssert interface {
	Assert(db *sql.DB) error
}

type Verify struct {
	Type  int // 1 = run at dml start.
	Sleep int
}

func LoadVerification(filePath string) ([]Verify, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		log.Println("load file error, ", filePath)
		return nil, err
	}

	defer func() {
		_ = jsonFile.Close()
	}()

	return nil, nil
}
