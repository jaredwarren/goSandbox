package game

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

	m.HandleFunc("/newgame/", common.MakeHandler(newgame, db, config)).Methods("GET")
	tmpl["newgame.html"] = template.Must(template.ParseFiles("static/templates/user/index.html", "static/templates/user/base.html"))
	return m
}
