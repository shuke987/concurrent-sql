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
	return nil
}
