package user

import (
	"database/sql"

	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"litmosauthor.com/unison/common"
	//"log"
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

func loginForm(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("User::Login")
	// for now parse every request so I don't have to recompile, maybe
	tmpl := make(map[string]*template.Template)
	tmpl["login.html"] = template.Must(template.ParseFiles("static/templates/user/index.html", "static/templates/user/base.html"))

	pagedata := &common.Page{Tags: &common.Tags{Id: 1, Name: "golang"},
		Content: &common.Content{Id: 9, Title: "Hello", Content: "World!"},
		Comment: &common.Comment{Id: 2, Note: "Good Day!"}}

	tmpl["login.html"].ExecuteTemplate(w, "base", pagedata)
}

func login(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("LOGIN....")
}

//var templates = template.Must(template.ParseFiles("static/templates/404.html", "static/templates/home.html"))

/*func executeTemplate(w http.ResponseWriter, tmpl string) {
	fmt.Printf("Project Template:" + tmpl)
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}*/
