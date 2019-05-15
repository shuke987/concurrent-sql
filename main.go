package main

import (
	"concurrent-sql/tests"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var paramDir = flag.String("dir", "test-cases", "specify the test case directory")

func main() {
	log.Printf("begin test")

	// 1. find all test cases.
	flag.Parse()
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
