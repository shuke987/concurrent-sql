package tests

import (
	"concurrent-sql/ddl"
	"concurrent-sql/dml"
	"concurrent-sql/verify"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
)

type TestCase struct {
	DSN           string
	DB            string
	DDL           ddl.DDL
	DML           []*dml.DML
	Verifications []verify.Verify
}

func (testCase *TestCase) Load(cfg *Config) error {
	testCase.DSN = cfg.DSN

	if err := testCase.DDL.Load(cfg.DDLFile); err != nil {
		return err
	}

	for i := range cfg.DMLFiles {
		d := &dml.DML{}
		if err := d.Load(cfg.DMLFiles[i]); err != nil {
			return err
		}
		d.Repeats = cfg.DMLRepeats[i]
		d.DSN = cfg.DMLdsn

		testCase.DML = append(testCase.DML, d)
	}

	if v, err := verify.LoadVerificationFromFile(cfg.VerificationFile); err != nil {
		return err
	} else {
		for i := range v {
			v[i].DSN = cfg.DMLdsn
		}
		testCase.Verifications = v
	}

	return nil
}

func (testCase *TestCase) Run() error {
	if err := testCase.runDDL(); err != nil {
		return err
	}

	if err := testCase.runDMLAndVerify(); err != nil {
		return err
	}

	if err := testCase.runAfterDML(); err != nil {
		return err
	}

	return nil
}

func (testCase *TestCase) runDDL() error {
	if ddlClient, err := sql.Open("mysql", testCase.DSN); err != nil {
		return err
	} else if err := ddlClient.Ping(); err != nil {
		log.Println("ping database error: ", testCase.DSN, ", ", err)
		return err
	} else {
		testCase.DDL.DB = ddlClient
	}

	defer func() {
		if err := testCase.DDL.DB.Close(); err != nil {
			// print error
			log.Println("close database error ", err)
		} else {
			testCase.DDL.DB = nil
		}
	}()

	if err := testCase.DDL.Run(); err != nil {
		return err
	}

	return nil
}

func (testCase *TestCase) runDMLAndVerify() (err error) {
	var chans []chan string
	shutdown := make(chan struct{})
	errorOccurs := false
	dmlCount := 0

	for i := 0; i < len(testCase.DML); i++ {
		ch := make(chan string)
		go testCase.DML[i].RunAsync(ch, shutdown)
		chans = append(chans, ch)
		dmlCount++
	}

	// run verify.
	for i := 0; i < len(testCase.Verifications); i++ {
		if testCase.Verifications[i].RunAt != "dml_end" {
			ch := make(chan string)
			go testCase.Verifications[i].RunAsync(ch, shutdown)
			chans = append(chans, ch)
		}
	}

	cases := make([]reflect.SelectCase, len(chans))
	for i, ch := range chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	remaining := len(cases)
	closeAlready := false

	for remaining > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			cases[chosen].Chan = reflect.ValueOf(nil)
			remaining -= 1

			if chosen < dmlCount {
				for i := 0; i < dmlCount; i++ {
					if cases[i].Chan != reflect.ValueOf(nil) {
						break
					} else if i == dmlCount-1 {
						if !closeAlready {
							close(shutdown)
							closeAlready = true
						}
					}
				}
			}

			continue
		} else {
			if v := value.String(); v != "" {
				log.Println("error occurs. verify id: ", chosen, ", ", v)

				// notify all go routines to quit.
				if !errorOccurs {
					errorOccurs = true
					err = errors.New(v)
					if !closeAlready {
						close(shutdown)
						closeAlready = true
					}
				}
			}
		}
	}

	if errorOccurs {
		testCase.afterFail()
	}

	return
}

func (testCase *TestCase) runAfterDML() (err error) {
	var chans []chan string
	shutdown := make(chan struct{})
	errorOccurs := false

	// run verify.
	for i := 0; i < len(testCase.Verifications); i++ {
		if testCase.Verifications[i].RunAt == "dml_end" {
			ch := make(chan string)
			go testCase.Verifications[i].RunAsync(ch, shutdown)
			chans = append(chans, ch)
		}
	}

	cases := make([]reflect.SelectCase, len(chans))
	for i, ch := range chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	remaining := len(cases)
	for remaining > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			cases[chosen].Chan = reflect.ValueOf(nil)
			remaining -= 1
			continue
		} else {
			if v := value.String(); v != "" {
				errStr := fmt.Sprintf("error occurs. verify id: %d. err=%s", chosen, v)
				log.Println(errStr)
				err = errors.New(errStr)
				// notify all go routines to quit.
				if !errorOccurs {
					errorOccurs = true
					close(shutdown)
				}
			}
		}
	}

	if errorOccurs {
		testCase.afterFail()
	}

	return
}

func (testCase *TestCase) afterFail() {
	// kill local tidb.
}
