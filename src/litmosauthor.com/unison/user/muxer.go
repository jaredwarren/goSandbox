package user

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"html/template"
	"litmosauthor.com/unison/common"
	"net/http"
)

var tmpl = make(map[string]*template.Template)

func MakeMuxer(prefix string, db *sql.DB) http.Handler {
	var m *mux.Router
	if prefix == "" {
		m = mux.NewRouter()
	} else {
		m = mux.NewRouter().PathPrefix(prefix).Subrouter()
	}

	// login form
	m.HandleFunc("/login/", common.MakeHandler(loginForm, db)).Methods("GET")
	m.HandleFunc("/login", common.MakeHandler(loginForm, db)).Methods("GET")
	tmpl["login.html"] = template.Must(template.ParseFiles("static/templates/user/index.html", "static/templates/user/base.html"))

	// login/logout
	m.HandleFunc("/login/", common.MakeHandler(login, db)).Methods("POST")
	m.HandleFunc("/logout/", common.MakeHandler(logout, db)).Methods("GET")

	// dashboard
	m.HandleFunc("/dashboard/", common.MakeHandler(dashboard, db)).Methods("GET")
	tmpl["dashboard.html"] = template.Must(template.ParseFiles("static/templates/user/dashboard/index.html", "static/templates/user/base.html"))
	return m
}
