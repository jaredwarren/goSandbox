package common

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"strings"
	"youtube/ini"
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

func MakeHandler(fn func(http.ResponseWriter, *http.Request, *sql.DB, *ini.Dict), db *sql.DB, config *ini.Dict) http.HandlerFunc {
	// do some checking, e.g. db - alpha or beta
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db, config)
	}
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]
	fmt.Println("Path:", path)
	http.FileServer(http.Dir("./"))

}

func NotFoundFunc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["path"]
	fmt.Println("Project::NotFoundFunc - " + category)
	fmt.Printf("404 - %s", category)
}

func GetCustId(w http.ResponseWriter, r *http.Request) string {
	//The Host that the user queried.
	host := r.URL.Host
	//host := "asdf.litmosauthor.com"
	//fmt.Printf("%+v\n", host)
	host = strings.TrimSpace(host)
	//Figure out if a subdomain exists in the host given.
	host_parts := strings.Split(host, ".")
	if len(host_parts) > 2 {
		//The subdomain exists, we store it as the first element
		//in a new array
		//subdomain := []string{host_parts[0]}
		return host_parts[0]
	}
	//http.Error(w, err.Error(), http.StatusInternalServerError)
	return "none"
}

var templates = template.Must(template.ParseFiles("static/templates/404.html", "static/templates/home.html"))

func executeTemplate(w http.ResponseWriter, tmpl string) {
	fmt.Printf("Project Template:" + tmpl)
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
