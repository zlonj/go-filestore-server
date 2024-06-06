package mysql

import (
	"database/sql"
	"fmt"
	"os"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	fmt.Println("Initing db connection to mysql server at port 3306...")
	db, _ = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/fileserver?charset=utf8")
	db.SetMaxOpenConns(1000)
	err := db.Ping()
	if err != nil {
		fmt.Println("Failed to connect to mysql, err: " + err.Error())
		os.Exit(1)
	}
}

func DBConn() *sql.DB {
	return db
}

func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()

	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
			scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
			err := rows.Scan(scanArgs...)
			checkErr(err)

			for i, col := range values {
					if col != nil {
							record[columns[i]] = col
					}
			}
			records = append(records, record)
	}
	return records
}

func checkErr(err error) {
if err != nil {
	log.Fatal(err)
	panic(err)
}
}
