package common

import (
	"fmt"
	//"github.com/gorilla/mux"
	"html/template"
	"net/http"
	//"strings"
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

func MakeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	// do some checking, e.g. db - alpha or beta
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
	}
}

var templates = template.Must(template.ParseFiles("static/templates/404.html", "static/templates/home.html"))

func executeTemplate(w http.ResponseWriter, tmpl string) {
	fmt.Printf("Project Template:" + tmpl)
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
