package main

import (
	"concurrent-sql/tests"
	"concurrent-sql/verify"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

var paramDir = flag.String("dir", "test-cases", "specify the test case directory")
var genExpect = flag.Bool("gen", false, "generate a expect result of specified query")
var dsn = flag.String("dsn", "root@tcp(127.0.0.1:4000)/?allowNativePasswords=true&maxAllowedPacket=0", "db connection")
var query = flag.String("query", "", "specify the query to be execute to get the expect result string")

func main() {

	// 1. find all test cases.
	flag.Parse()
	if *genExpect {
		printExpectResult(*dsn, *query)
		return
	}

	log.Printf("begin test")
	log.Printf("dir=%s", *paramDir)

	var testCases []*tests.TestCase
	if cases, err := tests.LoadCases(*paramDir); err != nil {
		log.Fatal(err)
	} else {
		testCases = cases
		log.Printf("%d cases loaded", len(testCases))
	}

	// 2. invoke each case's run.
	for _, c := range testCases {
		if err := c.Run(); err != nil {
			// error happened.
			log.Fatal(err)
		}
	}

	log.Printf("test finish")
}

func printExpectResult(dsn, query string) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("connect to db failed", err)
	}
	result, err := verify.GetQueryResult(db, query)
	if err != nil {
		log.Fatal("query from failed", err)
	}
	str := result.ToOneString()
	str = strings.ReplaceAll(str, "\n", "\\n")
	str = strings.ReplaceAll(str, "\t", "\\t")
	fmt.Println(str)
}
