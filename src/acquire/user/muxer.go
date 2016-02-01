package user

import (
	"acquire/common"
	"acquire/ini"
	"database/sql"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
)

var tmpl = make(map[string]*template.Template)

func MakeMuxer(prefix string, db *sql.DB, config *ini.Dict) http.Handler {
	var m *mux.Router
	if prefix == "" {
		m = mux.NewRouter()
	} else {
		m = mux.NewRouter().PathPrefix(prefix).Subrouter()
	}

	// login form
	m.HandleFunc("/login/", common.MakeHandler(loginForm, db, config)).Methods("GET")
	m.HandleFunc("/login", common.MakeHandler(loginForm, db, config)).Methods("GET")
	tmpl["login.html"] = template.Must(template.ParseFiles("static/templates/user/index.html", "static/templates/user/base.html"))

	// login/logout
	m.HandleFunc("/login/", common.MakeHandler(login, db, config)).Methods("POST")
	m.HandleFunc("/logout/", common.MakeHandler(logout, db, config)).Methods("GET")

	// dashboard
	m.HandleFunc("/dashboard/", common.MakeHandler(dashboard, db, config)).Methods("GET")
	tmpl["dashboard.html"] = template.Must(template.ParseFiles("static/templates/user/dashboard/index.html", "static/templates/user/base.html"))
	return m
}
