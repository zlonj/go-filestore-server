package handler

import (
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	PASSWORD_SALT = "*#890"
	TOKEN_SALT = "_tokensalt"
)


// Handles user sign up requests
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	username := r.Form.Get("username")
	password := r.Form.Get("password")

	if len(username) < 3 || len(password) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}

	encryptedPassword := util.Sha1([]byte(password + PASSWORD_SALT))
	success := dblayer.UserSignup(username, encryptedPassword)
	if success {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/static/view/signin.html", http.StatusNotFound)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	encryptedPassword := util.Sha1([]byte(password + PASSWORD_SALT))
	// 1. Verify username and password
	pwdChecked := dblayer.UserSignin(username, encryptedPassword)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}

	// 2. Generate token
	token := GenerateToken(username)
	updatesToken := dblayer.UpdatToken(username, token)
	if !updatesToken {
		w.Write([]byte("FAILED"))
		return
	}
	// 3. Redirect to home page after successful login
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request params
	r.ParseForm()
	username := r.Form.Get("username")

	// 2. Verify token (logic moved to auth:HTTPInterceptor)

	// 3. Get user information
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. Construct response body
	response := util.RespMsg{
		Code: 0,
		Msg: "OK",
		Data: user,
	}
	w.Write(response.JSONBytes())
}

func GenerateToken(username string) string {
	// 40 bites token: md5(username + timestamp + token_salt) + timestamp[:8]
	timestamp := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + timestamp + TOKEN_SALT))
	return tokenPrefix + timestamp[:8]
}

func IsTokenValid(token string) bool {
	return len(token) == 40
}
