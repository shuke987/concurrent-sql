package verify

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

const RUN_ONETIME = "dml_end"
const ASSERT_TYPE_ADMIN = "admin_check"

type SQLAssert interface {
	Assert(db *sql.DB) error
}

type Verify struct {
	RunAt   string   `json:"run_at"`
	Sleep   int      `json:"wait,omitempty"`
	Asserts []Assert `json:"asserts,omitempty"`
	DSN     string   `json:"-"`
}

func (v *Verify) RunAsync(c chan string, shutdown chan struct{}) {
	defer close(c)
	db, err := sql.Open("mysql", v.DSN)
	if err != nil {
		c <- fmt.Sprintf("%v", err)
		return
	}
	defer func() {
		_ = db.Close()
	}()

	for {
		select {
		case msg1 := <-shutdown:
			{
				fmt.Println("shutdown signal received", msg1)
				return
			}
		default:
			fmt.Println("no shutdown signal")
		}

		fmt.Println("start to execute verify case")
		err := v.Assert(db)
		if err != nil {
			c <- fmt.Sprintf("%v", err)
			return
		}
		if v.RunAt == RUN_ONETIME {
			return
		}
		fmt.Printf("execute done, sleep, %d", v.Sleep)
		time.Sleep(time.Duration(v.Sleep) * time.Second)
	}
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
		queryResult, err := getAllRecordAsString(db, as.SQL)
		if err != nil {
			return err
		}
		if as.Type == ASSERT_TYPE_ADMIN {
			fmt.Println("admin check without error")
			continue
		} else {
			equals := true
			if queryResult != as.Expect {
				fmt.Printf("Result is not equals to Expect:\nExpect is:%s\n Actually Result is:\n%s\n", as.Expect, queryResult)
				equals = false
				//now adjust
				for _, adjust := range as.Adjust {
					fmt.Printf("try to adjust sql: %s\n", adjust)
					_, err := db.Exec(adjust)
					if err != nil {
						fmt.Printf("execute adjust failed\n")
						return err
					}
					// check again
					queryResult, err = getAllRecordAsString(db, as.SQL)
					if err != nil {
						return err
					}
					if queryResult == as.Expect {
						equals = true
						break
					} else {
						fmt.Printf("Result is not equals to Expect:\nExpect is:%s\n Actually Result is:\n%s\n", as.Expect, queryResult)
					}
				}
			}
			if !equals {
				fmt.Println("the sql result not equals")
				return errors.New("verify case failed")
			}
		}
	}
	return nil
}

type SqlQueryResult struct {
	data        [][][]byte
	header      []string
	columnTypes []*sql.ColumnType
}

//append all rows to one string, rows are split by \n and columns are split by \t
func (result *SqlQueryResult) ToOneString() string {
	if result.data == nil || result.header == nil {
		return "no result"
	}
	var buf bytes.Buffer
	for rowIndex, row := range result.data {
		for index, col := range row {

			buf.WriteString(string(col))
			if index < len(row)-1 {
				buf.WriteString("\t")
			}
		}
		if rowIndex < len(result.data)-1 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

func getAllRecordAsString(db *sql.DB, query string) (string, error) {
	var str string
	queryResult, err := getQueryResult(db, query)
	if err != nil {
		return str, err
	}
	return queryResult.ToOneString(), nil
}

//get the query result
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

// readable query result like mysql shell client
func (result *SqlQueryResult) String() string {
	if result.data == nil || result.header == nil {
		return "no result"
	}

	// Calculate the max column length
	var colLength []int
	for _, c := range result.header {
		colLength = append(colLength, len(c))
	}
	for _, row := range result.data {
		for n, col := range row {
			if l := len(col); colLength[n] < l {
				colLength[n] = l
			}
		}
	}
	// The total length
	var total = len(result.header) - 1
	for index := range colLength {
		colLength[index] += 2 // Value will wrap with space
		total += colLength[index]
	}

	var lines []string
	var push = func(line string) {
		lines = append(lines, line)
	}

	// Write table header
	var header string
	for index, col := range result.header {
		length := colLength[index]
		padding := length - 1 - len(col)
		if index == 0 {
			header += "|"
		}
		header += " " + col + strings.Repeat(" ", padding) + "|"
	}
	splitLine := "+" + strings.Repeat("-", total) + "+"
	push(splitLine)
	push(header)
	push(splitLine)

	// Write rows data
	for _, row := range result.data {
		var line string
		for index, col := range row {
			length := colLength[index]
			padding := length - 1 - len(col)
			if index == 0 {
				line += "|"
			}
			line += " " + string(col) + strings.Repeat(" ", padding) + "|"
		}
		push(line)
	}
	push(splitLine)
	return strings.Join(lines, "\n")
}
