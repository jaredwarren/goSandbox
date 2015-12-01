package project

import (
	"database/sql"
	//"fmt"
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
	m.HandleFunc("/", common.MakeHandler(Dashboard, db))

	// setup templates here....
	tmpl["index.html"] = template.Must(template.ParseFiles("static/templates/index.html", "static/templates/content.html", "static/templates/base.html"))
	tmpl["other.html"] = template.Must(template.ParseFiles("static/templates/other.html", "static/templates/base.html"))

	//m.HandleFunc("/{path:.*}", common.MakeHandler(common.NotFoundFunc, db))
	return m
}
