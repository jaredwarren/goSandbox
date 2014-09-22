package project

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
)

type Tags struct {
	Id   int
	Name string
}

type Content struct {
	Id      int
	Title   string
	Content string
}

type Comment struct {
	Id   int
	Note string
}

type Page struct {
	Tags    *Tags
	Content *Content
	Comment *Comment
}

type Project struct {
}

func NewProject() *Project {
	return &Project{}
}

func Dashboard(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("Project::Dashboard")
	cust_id := "unison"
	rows, err := db.Query("SELECT project_name FROM project WHERE cust_id=?", cust_id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("~~%s~~\n", name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	//executeTemplate(w, "home")
	tmpl := make(map[string]*template.Template)
	tmpl["index.html"] = template.Must(template.ParseFiles("static/templates/index.html", "static/templates/base.html"))
	tmpl["other.html"] = template.Must(template.ParseFiles("static/templates/other.html", "static/templates/base.html"))

	pagedata := &Page{Tags: &Tags{Id: 1, Name: "golang"},
		Content: &Content{Id: 9, Title: "Hello", Content: "World!"},
		Comment: &Comment{Id: 2, Note: "Good Day!"}}

	tmpl["index.html"].ExecuteTemplate(w, "base", pagedata)
}

var templates = template.Must(template.ParseFiles("static/templates/404.html", "static/templates/home.html"))

func executeTemplate(w http.ResponseWriter, tmpl string) {
	fmt.Printf("Project Template:" + tmpl)
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
