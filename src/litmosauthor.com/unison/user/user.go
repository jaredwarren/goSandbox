package user

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"litmosauthor.com/unison/common"
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
	rows, err := db.Query("SELECT user_id, login, password, fullname, cust_id FROM user WHERE login=? LIMIT 1", username)

	if err != nil {
		return user, err
	}
	defer rows.Close()
	for rows.Next() {
		user = &User{}
		if err := rows.Scan(&user.Id, &user.Username, &user.Password, &user.Name, &user.CustId); err != nil {
			return user, err
		}
		//projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return user, err
	}
	if user == nil {
		return user, errors.New("Can't find user")
	}

	return user, nil
}

func loginForm(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("User::Login")
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
		// todo throw error
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Invalid User")
		//http.Error(w, "Invalid User", 401)
		return
	}

	// TOOD redirect
	fmt.Println(user)
}

//var templates = template.Must(template.ParseFiles("static/templates/404.html", "static/templates/home.html"))

/*func executeTemplate(w http.ResponseWriter, tmpl string) {
	fmt.Printf("Project Template:" + tmpl)
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}*/
