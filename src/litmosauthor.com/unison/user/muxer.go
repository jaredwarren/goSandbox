package project

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"net/http"
)

func MakeMuxer(prefix string, db *sql.DB) http.Handler {
	var m *mux.Router
	if prefix == "" {
		m = mux.NewRouter()
	} else {
		m = mux.NewRouter().PathPrefix(prefix).Subrouter()
	}
	m.HandleFunc("/login", makeHandler(loginForm, db)).Methods("GET")
	m.HandleFunc("/login", makeHandler(login, db)).Methods("POST")
	m.HandleFunc("/{path:.*}", makeHandler(NotFoundFunc, db))
	return m
}

func loginForm(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	category := vars["path"]
	fmt.Println("Project::NotFoundFunc - " + category)
	fmt.Printf("404 - %s", category)
}

func login(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	category := vars["path"]
	fmt.Println("Project::NotFoundFunc - " + category)
	fmt.Printf("404 - %s", category)
}

/*func makeHandler(fn func(http.ResponseWriter, *http.Request, *sql.DB), db *sql.DB) http.HandlerFunc {
	// do some checking, e.g. db - alpha or beta
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, db)
	}
}*/

func NotFoundFunc(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	category := vars["path"]
	fmt.Println("Project::NotFoundFunc - " + category)
	fmt.Printf("404 - %s", category)
}
