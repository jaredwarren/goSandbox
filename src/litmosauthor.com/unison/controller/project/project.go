package controller

import (
	"database/sql"
	"fmt"
	//"html/template"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

type ProjectController struct {
}

func NewProjectController() *ProjectController {
	return &ProjectController{}
}

func (this *ProjectController) Dashboard(res http.ResponseWriter, req *http.Request, db *sql.DB) {
	fmt.Println("home")
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

	//p := &Page{Title: "title", Body: []byte("body")}
	this.executeTemplate(res, "home")
}

func (this *ProjectController) executeTemplate(w http.ResponseWriter, tmpl string) {
	fmt.Printf("Project Template:" + tmpl)
	//err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	//if err != nil {
	//    http.Error(w, err.Error(), http.StatusInternalServerError)
	//}
}
