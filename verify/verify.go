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

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	RUN_ONETIME       = "dml_end"
	ASSERT_TYPE_ADMIN = "admin_check"
	ASSERT_TYPE_PLAN  = "plan"
)

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
				log.Println("shutdown signal received", msg1)
				return
			}
		default:
			log.Println("no shutdown signal")
		}

		log.Println("start to execute verify case")
		err := v.Assert(db)
		if err != nil {
			c <- fmt.Sprintf("%v", err)
			return
		}
		if v.RunAt == RUN_ONETIME {
			return
		}
		log.Printf("execute done, sleep, %d", v.Sleep)
		time.Sleep(time.Duration(v.Sleep) * time.Second)
	}
}

type Assert struct {
	Type   string   `json:"type,omitempty"`
	SQL    string   `json:"sql,omitempty"`
	Adjust []string `json:"adjust,omitempty"`
	Expect string   `json:"expect,omitempty"`
	Clean  []string `json:"clean,omitempty"`
}

//clean assert variable data
func (assert *Assert) CleanEnv(db *sql.DB) {
	for _, query := range assert.Clean {
		_, err := GetQueryResult(db, query)
		if err != nil {
			log.Println("execute clean sql failed! ", query)
		}
	}
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
		queryResult, err := GetQueryResult(db, as.SQL)
		if err != nil {
			return err
		}
		switch as.Type {
		case ASSERT_TYPE_ADMIN:
			log.Println("admin check without error")
		default:
			stringFunc := queryResult.getQueryResultStringFunc(as.Type)
			queryResultStr := stringFunc()
			equals := true
			if queryResultStr != as.Expect {
				fmt.Println("Result is not equals to Expect")
				printDiff(as.Expect, queryResultStr)
				equals = false
				//now adjust
				for _, adjust := range as.Adjust {
					log.Printf("try to adjust sql: %s\n", adjust)
					_, err := db.Exec(adjust)
					if err != nil {
						log.Printf("execute adjust failed\n")
						return err
					}
					// check again

					queryResult, err := GetQueryResult(db, as.SQL)
					if err != nil {
						return err
					}
					stringFunc := queryResult.getQueryResultStringFunc(as.Type)
					queryResultStr = stringFunc()
					if queryResultStr == as.Expect {
						equals = true
						break
					} else {
						log.Println("Result is not equals to Expect")
						printDiff(as.Expect, queryResultStr)
					}
				}
			}

			//let's clean env first
			as.CleanEnv(db)
			if !equals {
				fmt.Println("the sql result not equals")
				return errors.New("verify case failed")
			} else {
				log.Println("plan assert successfully!")
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
	queryResult, err := GetQueryResult(db, query)
	if err != nil {
		return str, err
	}
	return queryResult.ToOneString(), nil
}

//get the query result
func GetQueryResult(db *sql.DB, query string) (*SqlQueryResult, error) {
	log.Println("executing sql", query)
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

func printDiff(expect, actual string) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	patch := diffmatchpatch.New()
	diff := patch.DiffMain(expect, actual, false)
	var newExpectedContent, newActualResult bytes.Buffer
	for _, d := range diff {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			newExpectedContent.WriteString(d.Text)
			newActualResult.WriteString(d.Text)
		case diffmatchpatch.DiffDelete:
			newExpectedContent.WriteString(red(d.Text))
		case diffmatchpatch.DiffInsert:
			newActualResult.WriteString(green(d.Text))
		}
	}
	fmt.Printf("Expected Result:\n%s\nActual Result:\n%s\n", newExpectedContent.String(), newActualResult.String())
}

func (result *SqlQueryResult) getPlanScanType() string {
	if result.data == nil || result.header == nil {
		return "no result"
	}
	for _, row := range result.data {
		firstCol := string(row[0])
		if strings.Contains(firstCol, "IndexScan") {
			return "IndexScan"
		} else if strings.Contains(firstCol, "TableScan") {
			return "TableScan"
		}
	}
	return ""
}

func (result *SqlQueryResult) getQueryResultStringFunc(compareType string) func() string {
	switch compareType {
	case ASSERT_TYPE_PLAN:
		return result.getPlanScanType
	default:
		return result.ToOneString
	}
}
