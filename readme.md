## test database with sqls

### usage
- clone project
- write case.
- ./concurrent-sql -dir=/path_to_case

### case sample

    [Global]
    dsn=root@tcp(127.0.0.1:4000)/?allowNativePasswords=true&maxAllowedPacket=0
    [DDL]
    file=ddl.sql
    [DML]
    dsn=root@tcp(127.0.0.1:4000)/test2?allowNativePasswords=true&maxAllowedPacket=0
    file=dml-1.sql,1
    file2=dml-2.sql,1
    [Verify]
    verify=verification.json

ddl: sqls to init database and tables

dml section: dml files with sqls to run, and how many times it will repeat. 

At least you need one dml file. Otherwise nothing is done.

verify: 

    [
      {
        "run_at": "dml_start", // run unil dml is finished.
        "wait": 0, // interval between two repeat run.
        "asserts":[
          {
            "type": "admin_check", // only checks whether there is an error
            "sql": "select * from mysql.user;",
            "adjust":[],
            "expect": "xxx"
          }
        ]
      },
      {
        "run_at": "dml_end",
        "wait": 10,
        "asserts": [
          {
            "type": "plan", // check result == expect.
            "sql": "explain select * from mysql.user; ",
            "adjust":["select * from mysql.user;", "select * from mysql.user;"],
            "expect": "xxx"
          }
        ]
      }
    ]
    
### more case
in ./test-cases