package builder

import (
	"builder/common"
	"builder/ini"
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

	// Read
	m.HandleFunc("/read/{type}", common.MakeHandler(Read, db, config)).Methods("GET")
	// Create
	m.HandleFunc("/create/{type}", common.MakeHandler(Create, db, config)).Methods("POST")
	// Destroy
	m.HandleFunc("/destroy/{type}", common.MakeHandler(Destroy, db, config)).Methods("GET")
	m.HandleFunc("/destroy/{type}", common.MakeHandler(Destroy, db, config)).Methods("DELETE")
	// Update
	m.HandleFunc("/update/{type}", common.MakeHandler(Update, db, config)).Methods("PUT")
	m.HandleFunc("/update/{type}", common.MakeHandler(Update, db, config)).Methods("POST")

	// Save
	m.HandleFunc("/save/", common.MakeHandler(Save, db, config)).Methods("POST")
	m.HandleFunc("/saveBackup/", common.MakeHandler(SaveBackup, db, config)).Methods("POST")
	m.HandleFunc("/saveSettings/", common.MakeHandler(SaveSettings, db, config)).Methods("POST")

	tmpl["index.html"] = template.Must(template.ParseFiles("static/templates/watch/index.html", "static/templates/watch/base.html"))
	return m
}
