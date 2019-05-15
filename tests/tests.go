package tests

import (
	"log"
)

func LoadCases(dir string) ([]*TestCase, error) {
	configFiles, err := findAllConfigs(dir)
	if err != nil {
		return nil, err
	}

	log.Println(len(configFiles), " found")

	var testCases []*TestCase

	for _, o := range configFiles {
		cfg := &Config{}
		if err := cfg.Load(o); err != nil {
			return nil, err
		}

		c := &TestCase{}
		if err := c.Load(cfg); err != nil {
			return nil, err
		} else {
			testCases = append(testCases, c)
		}
	}

	return testCases, nil
}
