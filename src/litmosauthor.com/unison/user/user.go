package project

import (
	"database/sql"

	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
	//"reflect"
)

type Users []User

type User struct {
	Id       string `db:"user_id"`
	Username string `db:"login"`
	Password string `db:"password"`
	Name     string `db:"fullname"`
	CustId   string `db:"cust_id"`
}

func Login(db *sql.DB) {
	projects := Projects{}

	cust_id := "unison"
	rows, err := db.Query("SELECT project_id, project_name, cust_id FROM project WHERE cust_id=?", cust_id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		//var project Project
		project := Project{}
		if err := rows.Scan(&project.Id, &project.Name, &project.CustId); err != nil {
			log.Fatal(err)
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return projects
}

func Dashboard(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("Project::Dashboard")

	projects := getProjects(db)
	//fmt.Println(projects.HasProjects())

	// for now parse every request so I don't have to recompile, maybe
	tmpl := make(map[string]*template.Template)
	tmpl["index.html"] = template.Must(template.ParseFiles("static/templates/index.html", "static/templates/content.html", "static/templates/base.html"))
	tmpl["other.html"] = template.Must(template.ParseFiles("static/templates/other.html", "static/templates/base.html"))

	pagedata := &Page{Tags: &Tags{Id: 1, Name: "golang"},
		Content:  &Content{Id: 9, Title: "Hello", Content: "World!"},
		Projects: projects,
		Comment:  &Comment{Id: 2, Note: "Good Day!"}}

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
