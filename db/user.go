package db

import (
	mydb "filestore-server/db/mysql"
	"fmt"
)

// Sign up user with username and password, persist into DB
func UserSignup(username string, password string) bool {
	stmt, err := mydb.DBConn().Prepare("INSERT IGNORE INTO table_user (`user_name`, `user_pwd`) values (?, ?)")
	if err != nil {
		fmt.Println("Failed to prepare insert statement, error: ", err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, password)
	if err != nil {
		fmt.Println("Failed to execute insert statement, error: " + err.Error())
		return false
	}

	if rowsAffected, err := ret.RowsAffected(); err == nil && rowsAffected > 0 {
		return true
	}
	return false
}

// Handles user sign in, returns whether username and password are found in DB
func UserSignin(username string, encryptedPassword string) bool {
	stmt, err := mydb.DBConn().Prepare("SELECT * FROM table_user WHERE user_name = ? LIMIT 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("username not found: " + username)
		return false
	}

	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encryptedPassword {
		return true
	}
	return false
}

// Updates token in DB for a user
func UpdatToken(username string, token string) bool {
	stmt, err := mydb.DBConn().Prepare("REPLACE INTO table_user_token (`user_name`, `user_token`) VALUES (?, ?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

type User struct {
	Username string
	Email string
	Phone string
	SignupAt string
	LastActiveAt string
	Status int
}

func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mydb.DBConn().Prepare(
		"SELECT user_name, signup_at FROM table_user WHERE user_name = ? LIMIT 1",
	)
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	return user, nil
}