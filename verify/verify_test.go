package verify

import (
	"testing"
)

func TestLoadVerificationFromData(t *testing.T) {
	var jsonData = `
[
  {
    "run_at": "dml_start",
    "wait": 0,
    "asserts":[
      {
        "type": "admin_check",
        "sql": "",
        "adjust":["sss", ""],
        "expect": "xxx"
      }
    ]
  },
  {
    "run_at": "dml_end",
    "wait": 10,
    "asserts": [
      {
        "type": "plan",
        "sql": "Explain ",
        "adjust":["sss", ""],
        "expect": ""
      }
    ]
  }
]`
	verifies, err := LoadVerificationFromData([]byte(jsonData))

	if err != nil {
		t.Fatalf("parse failed: err=%verifies", err)
	}
	if len(verifies) != 2 {
		t.Fatalf("size not 2: size=%d", len(verifies))
	}
}
