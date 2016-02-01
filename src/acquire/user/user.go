package user

import (
	"acquire/ini"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
	"litmosauthor.com/unison/common"
	"log"
	"net/http"
)

type Users []User

type User struct {
	Id           string `db:"user_id"`
	Username     string `db:"login"`
	Password     string `db:"password"`
	PasswordHash []byte `db:"password2"`
	Name         string `db:"fullname"`
	CustId       string `db:"cust_id"`
}

func LoginUser(username string, password string, db *sql.DB, config *ini.Dict) (user *User, err error) {
	// query db
	rows, err := db.Query("SELECT user_id, login, password, password2, fullname, cust_id FROM user WHERE login=? LIMIT 1", username)
	if err != nil {
		log.Fatal(err)
		return user, err
	}
	defer rows.Close()
	// populate new user
	for rows.Next() {
		user = &User{}
		if err := rows.Scan(&user.Id, &user.Username, &user.Password, &user.PasswordHash, &user.Name, &user.CustId); err != nil {
			fmt.Println("query scan error")
			return user, errors.New("Can't find user")
		}
	}
	if err := rows.Err(); err != nil {
		fmt.Println("rows error")
		return user, errors.New("Can't find user")
	}
	if user == nil {
		return user, errors.New("Can't find user")
	}
	// hash password
	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		return user, errors.New("Invalid Password")
	}

	return user, nil
}

func setSession(userName User, w http.ResponseWriter) {
	value := map[string]User{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	}
}

var cookieHandler = securecookie.New(securecookie.GenerateRandomKey(64), securecookie.GenerateRandomKey(32))

func GetUserName(r *http.Request) (userName *User) {
	if cookie, err := r.Cookie("session"); err == nil {
		cookieValue := make(map[string]*User)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func loginForm(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	fmt.Println("User::Login")
	userName := GetUserName(r)
	if userName != nil {
		http.Redirect(w, r, "/user/dashboard/", 302)
		return
	}

	pagedata := &common.Page{Tags: &common.Tags{Id: 1, Name: "golang"},
		Content: &common.Content{Id: 9, Title: "Hello", Content: "World!"},
		Comment: &common.Comment{Id: 2, Note: "Good Day!"}}

	tmpl["login.html"].ExecuteTemplate(w, "base", pagedata)
}

func login(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := LoginUser(username, password, db, config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	setSession(*user, w)

	// TODO: create response struct and json that.
	json, _ := json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func logout(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	clearSession(w)
	http.Redirect(w, r, "/user/login/", 302)
}

func clearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

func dashboard(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	fmt.Println("User::Dishboard")
	userName := GetUserName(r)
	if userName == nil {
		http.Redirect(w, r, "/user/login/", 302)
		return
	}

	pagedata := &common.Page{
		Tags:    &common.Tags{Id: 1, Name: "golang"},
		Content: &common.Content{Id: 9, Title: "Hello", Content: "World!"},
		Comment: &common.Comment{Id: 2, Note: "Good Day!"},
	}

	tmpl["dashboard.html"].ExecuteTemplate(w, "base", pagedata)
}
