package verify

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
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

func TestSqlQueryResult_ToOneString(t *testing.T) {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:4000)/?allowNativePasswords=true&maxAllowedPacket=0")
	if err != nil {
		fmt.Printf("err=%v", err)
		return
	}
	result, err := getQueryResult(db, "explain select * from mysql.user;")
	if err != nil {
		fmt.Printf("err=%v", err)
		return
	}
	fmt.Println(result.ToOneString())
}

func TestVerify_Assert(t *testing.T) {
	vs := `[
  {
    "run_at": "dml_start",
    "wait": 0,
    "asserts":[
      {
        "type": "explain",
        "sql": "explain select * from mysql.user;",
        "adjust":["select * from mysql.user", "select * from mysql.user" ],
        "expect": "xxx"
      }
    ]
  }
]`

	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:4000)/?allowNativePasswords=true&maxAllowedPacket=0")
	if err != nil {
		fmt.Printf("err=%v", err)
		return
	}
	verifies, err := LoadVerificationFromData([]byte(vs))
	if err != nil {
		fmt.Printf("err=%v", err)
		return
	}
	for _, v := range verifies {
		v.Assert(db)
	}

}
