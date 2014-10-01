package common

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
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

func MakeHandler(fn func(http.ResponseWriter, *http.Request, *sql.DB), db *sql.DB) http.HandlerFunc {
	// do some checking, e.g. db - alpha or beta
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db)
	}
}

func NotFoundFunc(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	category := vars["path"]
	fmt.Println("Project::NotFoundFunc - " + category)
	fmt.Printf("404 - %s", category)
}
