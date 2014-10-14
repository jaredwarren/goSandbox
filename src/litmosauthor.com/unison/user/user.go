package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/securecookie"
	"html/template"
	"litmosauthor.com/unison/common"
	"log"
	"net/http"
)

type Users []User

type User struct {
	Id       string `db:"user_id"`
	Username string `db:"login"`
	Password string `db:"password"`
	Name     string `db:"fullname"`
	CustId   string `db:"cust_id"`
}

func LoginUser(username string, password string, db *sql.DB) (user *User, err error) {
	// TODO: hash password, and add to query

	// query db
	rows, err := db.Query("SELECT user_id, login, password, fullname, cust_id FROM user WHERE login=? LIMIT 1", username)
	if err != nil {
		log.Fatal(err)
		return user, err
	}
	defer rows.Close()
	// populate new user
	for rows.Next() {
		user = &User{}
		if err := rows.Scan(&user.Id, &user.Username, &user.Password, &user.Name, &user.CustId); err != nil {
			log.Fatal(err)
			return user, err
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
		return user, err
	}
	if user == nil {
		return user, errors.New("Can't find user")
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

func getUserName(r *http.Request) (userName *User) {
	if cookie, err := r.Cookie("session"); err == nil {
		cookieValue := make(map[string]*User)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func loginForm(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("User::Login")
	userName := getUserName(r)
	if userName != nil {
		http.Redirect(w, r, "/user/dashboard/", 302)
		return
	}

	// for now parse every request so I don't have to recompile, maybe
	//tmpl := make(map[string]*template.Template)
	tmpl := template.Must(template.ParseFiles("static/templates/user/index.html", "static/templates/user/base.html"))

	pagedata := &common.Page{Tags: &common.Tags{Id: 1, Name: "golang"},
		Content: &common.Content{Id: 9, Title: "Hello", Content: "World!"},
		Comment: &common.Comment{Id: 2, Note: "Good Day!"}}

	tmpl.ExecuteTemplate(w, "base", pagedata)
}

func login(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := LoginUser(username, password, db)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	setSession(*user, w)

	// TODO: create response struct and json that.
	json, _ := json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func logout(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	clearSession(w)
	http.Redirect(w, r, "/", 302)
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

func dashboard(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("User::Dishboard")
	userName := getUserName(r)
	if userName == nil {
		http.Redirect(w, r, "/user/login/", 302)
		return
	}

	// for now parse every request so I don't have to recompile, maybe
	tmpl := template.Must(template.ParseFiles("static/templates/user/dashboard/index.html", "static/templates/user/base.html"))

	pagedata := &common.Page{Tags: &common.Tags{Id: 1, Name: "golang"},
		Content: &common.Content{Id: 9, Title: "Hello", Content: "World!"},
		Comment: &common.Comment{Id: 2, Note: "Good Day!"}}

	tmpl.ExecuteTemplate(w, "base", pagedata)
}

func dashboardApp(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("User::Dishboard")
	userName := getUserName(r)
	if userName == nil {
		http.Redirect(w, r, "/user/login/", 302)
		return
	}

	// for now parse every request so I don't have to recompile, maybe
	// // TODO: create dasboard base
	tmpl := template.Must(template.ParseFiles("static/templates/user/dashboard/dashboard-app.html"))

	pagedata := &common.Page{Tags: &common.Tags{Id: 1, Name: "golang"},
		Content: &common.Content{Id: 9, Title: "Hello", Content: "World!"},
		Comment: &common.Comment{Id: 2, Note: "Good Day!"}}

	tmpl.ExecuteTemplate(w, "dashboardApp", pagedata)
}

//var templates = template.Must(template.ParseFiles("static/templates/404.html", "static/templates/home.html"))

/*func executeTemplate(w http.ResponseWriter, tmpl string) {
	fmt.Printf("Project Template:" + tmpl)
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}*/
