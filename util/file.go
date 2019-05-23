package util

import (
	"bufio"
	"io/ioutil"
	"os"

	"github.com/pingcap/log"
	"github.com/pingcap/parser"

	_ "github.com/pingcap/tidb/types/parser_driver"
)

func ReadFileLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lines := make([]string, 0, 10)

	reader := bufio.NewReader(f)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		lines = append(lines, string(line))
	}

	return lines, nil
}

func GetSQLStatements(path string) ([]string, error) {
	sqlBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p := parser.New()
	stmts, warns, err := p.Parse(string(sqlBytes), "", "")
	if err != nil {
		return nil, err
	}
	for _, w := range warns {
		log.Info("warn: " + w.Error())
	}

	lines := make([]string, 0, 10)
	for _, stmt := range stmts {
		lines = append(lines, stmt.Text())
	}
	return lines, nil
}
