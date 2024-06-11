package db

import (
	mydb "filestore-server/db/mysql"
	"fmt"
	"time"
)

type UserFile struct {
	UserName string
	FileHash string
	FileName string
	FileSize int64
	UploadAt string
	LastUpdate string
}

// Inserts user file record into user file table on file upload
func OnUserFileUploadFinished(
	username string,
	filehash string,
	filename string,
	filesize int64,
) bool {
	stmt, err := mydb.DBConn().Prepare(
		"INSERT IGNORE INTO table_user_file (`user_name`, `file_sha1`, `file_name`," +
		"`file_size`, `upload_at`) VALUES (?, ?, ?, ?, ?)",
	)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filehash, filename, filesize, time.Now())
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// Fetch at most `limit` user files from a user
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT file_sha1, file_name, file_size, upload_at, last_update FROM " +
		"table_user_file WHERE user_name = ? LIMIT ?",
	)
	if err != nil {
		fmt.Println("-1", err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, limit)
	if err != nil {
		fmt.Println("0", err.Error())
		return nil, err
	}

	var userFiles []UserFile
	for rows.Next() {
		userfile := UserFile{}
		err = rows.Scan(&userfile.FileHash, &userfile.FileName, &userfile.FileSize, &userfile.UploadAt, &userfile.LastUpdate)
		if err != nil {
			fmt.Println("1", err.Error())
			break
		}
		userFiles = append(userFiles, userfile)
	}
	return userFiles, nil
}