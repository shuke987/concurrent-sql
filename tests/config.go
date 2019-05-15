package tests

import (
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

type Config struct {
	DSN              string
	Database         string
	DDLFile          string
	DMLdsn           string
	DMLFiles         []string
	DMLRepeats       []int
	VerificationFile string
}

// find all case in dir and sub directories of dir, recursively.
func findAllConfigs(dir string) ([]string, error) {
	var configFiles []string

	// find case.ini in this directory.
	caseIni := path.Join(dir, "case.ini")
	if _, err := os.Stat(caseIni); err == nil {
		configFiles = append(configFiles, caseIni)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// find in sub directories.
	subDirs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range subDirs {
		if !file.IsDir() {
			continue
		}

		if subConfigFiles, err := findAllConfigs(path.Join(dir, file.Name())); err != nil {
			return nil, err
		} else {
			configFiles = append(configFiles, subConfigFiles...)
		}
	}

	return configFiles, nil
}

/*
 * sample format:

		[Global]
		dsn=root:/
		database=test
		[DDL]
		file=ddl.sql
		[DML]
		file=b.txt,1000
		file2=dml-2.sql,2000
		[Verify]
		query=query.json

*/
func (c *Config) Load(iniPath string) error {
	iniFile, err := ini.Load(iniPath)
	if err != nil {
		//log.Printf("load error: %s", configPath)
		return err
	}

	baseDir := path.Dir(iniPath)

	// global section
	c.DSN = iniFile.Section("Global").Key("dsn").String()
	c.Database = iniFile.Section("Global").Key("database").String()
	if c.DSN == "" || c.Database == "" {
		return errors.New("invalid dsn or database name")
	}

	// ddl section
	if ddlFile := iniFile.Section("DDL").Key("file").String(); ddlFile == "" {
		return errors.New("invalid ddl file name")
	} else {
		c.DDLFile = path.Join(baseDir, ddlFile)
	}

	// dml section
	c.DMLdsn = iniFile.Section("DML").Key("dsn").String()
	if c.DMLdsn == "" {
		return errors.New("empty dml dsn")
	}
	keys := iniFile.Section("DML").KeyStrings()
	for _, key := range keys {
		if key == "dsn" {
			continue
		}

		dmlConfig := iniFile.Section("DML").Key(key).String()
		if fileName, repeats, err := c.parseDML(dmlConfig); err != nil {
			return err
		} else {
			c.DMLFiles = append(c.DMLFiles, path.Join(baseDir, fileName))
			c.DMLRepeats = append(c.DMLRepeats, repeats)
		}
	}
	if len(c.DMLFiles) == 0 {
		return errors.New("invalid dml files")
	}

	// verify section
	if verifyFile := iniFile.Section("Verify").Key("verify").String(); verifyFile == "" {
		return errors.New(fmt.Sprintf("invalid verify file: %s", verifyFile))
	} else {
		c.VerificationFile = path.Join(baseDir, verifyFile)
	}

	return nil
}

// parse dml parameter into filename and repeat count.
// sample:  a.sql,1000 into fileName = a.sql, repeats=1000
func (c *Config) parseDML(line string) (fileName string, repeats int, err error) {
	params := strings.Split(line, ",")

	if len(params) == 1 {
		fileName = params[0]
		repeats = 1
	} else if len(params) == 2 {
		fileName = params[0]
		repeats, err = strconv.Atoi(params[1])
		if repeats <= 0 {
			repeats = 1
		}
	} else {
		err = errors.New("invalid dml parameter")
	}

	if err == nil && fileName == "" {
		err = errors.New("invalid dml file name")
	}

	return
}
