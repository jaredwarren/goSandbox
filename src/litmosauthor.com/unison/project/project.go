package project

import (
	"database/sql"

	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"litmosauthor.com/unison/common"
	"log"
	"net/http"
	//"reflect"
)

type Projects []Project

func (p Projects) HasProjects() bool {
	return len(p) > 0
}

type Project struct {
	Id     string `db:"project_id"`
	Name   string `db:"project_name"`
	CustId string `db:"cust_id"`
}

func getProjects(db *sql.DB) Projects {
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

type ProjectPage struct {
	Tags     *common.Tags
	Content  *common.Content
	Comment  *common.Comment
	Projects Projects
}

func Dashboard(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("Project::Dashboard")

	projects := getProjects(db)
	//fmt.Println(projects.HasProjects())

	// for now parse every request so I don't have to recompile, maybe
	tmpl := make(map[string]*template.Template)
	tmpl["index.html"] = template.Must(template.ParseFiles("static/templates/index.html", "static/templates/content.html", "static/templates/base.html"))
	tmpl["other.html"] = template.Must(template.ParseFiles("static/templates/other.html", "static/templates/base.html"))

	//fmt.Println(common.GetCustId(w, r))

	pagedata := &ProjectPage{Tags: &common.Tags{Id: 1, Name: "golang"},
		Content:  &common.Content{Id: 9, Title: "Hello", Content: "World!"},
		Projects: projects,
		Comment:  &common.Comment{Id: 2, Note: common.GetCustId(w, r)}}

	tmpl["index.html"].ExecuteTemplate(w, "base", pagedata)
}
