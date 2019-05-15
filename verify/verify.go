package verify

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
)

type SQLAssert interface {
	Assert(db *sql.DB) error
}

type Verify struct {
	RunAt   string   `json:"run_at"`
	Sleep   int      `json:"wait,omitempty"`
	Asserts []Assert `json:"asserts,omitempty"`
}

func (v *Verify) RunAsync(c chan string, shutdown chan struct{}) error {
	close(c)
	return nil
}

type Assert struct {
	Type   string   `json:"type,omitempty"`
	SQL    string   `json:"sql,omitempty"`
	Adjust []string `json:"adjust,omitempty"`
	Expect string   `json:"expect,omitempty"`
}

func LoadVerificationFromData(jsonData []byte) ([]Verify, error) {
	var verifies []Verify
	err := json.Unmarshal(jsonData, &verifies)
	return verifies, err
}

func LoadVerificationFromFile(filePath string) ([]Verify, error) {
	jsonData, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("load file error, ", filePath)
		return nil, err
	}
	return LoadVerificationFromData(jsonData)
}

func (verify *Verify) Assert(db *sql.DB) error {
	for _, as := range verify.Asserts {
		_, err := db.Exec(as.SQL)
		if err != nil {
			return err
		}

	}
	return nil
}

type SqlQueryResult struct {
	data        [][][]byte
	header      []string
	columnTypes []*sql.ColumnType
}

// readable query result like mysql shell client
func (result *SqlQueryResult) String() string {
	return ""
}

func getQueryResult(db *sql.DB, query string) (*SqlQueryResult, error) {
	result, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	cols, err := result.Columns()
	if err != nil {
		return nil, err
	}
	types, err := result.ColumnTypes()
	if err != nil {
		return nil, err
	}
	var allRows [][][]byte
	for result.Next() {
		var columns = make([][]byte, len(cols))
		var pointer = make([]interface{}, len(cols))
		for i := range columns {
			pointer[i] = &columns[i]
		}
		err := result.Scan(pointer...)
		if err != nil {
			return nil, err
		}
		allRows = append(allRows, columns)
	}
	queryResult := SqlQueryResult{data: allRows, header: cols, columnTypes: types}
	return &queryResult, nil
}
