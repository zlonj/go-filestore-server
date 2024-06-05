package db

import (
	mydb "filestore-server/db/mysql"
	"fmt"
)

func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string) bool {
	stmt, err := mydb.DbConn().Prepare(
		"insert ignore into tbl_file (`file_sha1`, `file_name`, `file_size`)" +
			"`file_addr`, `status` values (?,?,?,?,1)")
	if err != nil {
		fmt.Println("Failed to prepare sql statement, err: " + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if rowsAffected, err := ret.RowsAffected(); err == nil {
		if rowsAffected <= 0 {
			fmt.Printf("File with hash %s has been uploaded before\n", filehash)
		}
		return true
	}
	return false
}