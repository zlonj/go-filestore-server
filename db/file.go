package db

import (
	"database/sql"
	mydb "filestore-server/db/mysql"
	"fmt"
)

func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into table_file (`file_sha1`,`file_name`,`file_size`," +
			"`file_addr`,`status`) values (?,?,?,?,1)")
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

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// Get metadata from mysql db
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare("SELECT file_sha1, file_addr, file_name, file_size FROM table_file " +
		"WHERE file_sha1 = ? AND status = 1 LIMIT 1")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tableFile := TableFile {}
	err = stmt.QueryRow(filehash).Scan(&tableFile.FileHash, &tableFile.FileAddr, &tableFile.FileName, &tableFile.FileSize)
	if err != nil {
		if err == sql.ErrNoRows {
			// Found no rows, return both nil
			return nil, nil
		} else {
			fmt.Println(err.Error())
			return nil, err
		}
	}
	return &tableFile, nil
}