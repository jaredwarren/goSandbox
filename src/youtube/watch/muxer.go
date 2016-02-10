package watch

import (
	"database/sql"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"youtube/common"
	"youtube/ini"
)

var tmpl = make(map[string]*template.Template)

func MakeMuxer(prefix string, db *sql.DB, config *ini.Dict) http.Handler {
	var m *mux.Router
	if prefix == "" {
		m = mux.NewRouter()
	} else {
		m = mux.NewRouter().PathPrefix(prefix).Subrouter()
	}

	m.HandleFunc("/{tag}/", common.MakeHandler(All, db, config)).Methods("GET")

	tmpl["index.html"] = template.Must(template.ParseFiles("static/templates/watch/index.html", "static/templates/watch/base.html"))
	return m
}
