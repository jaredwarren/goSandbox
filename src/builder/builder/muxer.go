package builder

import (
	"builder/ini"
	"database/sql"
	"encoding/json"
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
	m.HandleFunc("/read/{projectId:[0-9]+}/{type}", MakeJsonHandler(Read, db, config)).Methods("GET")
	m.HandleFunc("/read/{projectId:[0-9]+}/{type}", MakeJsonHandler(ReadOptions, db, config)).Methods("OPTIONS")
	// Create
	m.HandleFunc("/create/{projectId:[0-9]+}/{type}", MakeJsonHandler(Create, db, config)).Methods("POST")
	m.HandleFunc("/create/{projectId:[0-9]+}/{type}", MakeJsonHandler(CreateOptions, db, config)).Methods("OPTIONS")
	// Destroy
	//m.HandleFunc("/destroy/{projectId:[0-9]+}/{type}", MakeJsonHandler(Destroy, db, config)).Methods("GET")
	m.HandleFunc("/destroy/{projectId:[0-9]+}/{type}", MakeJsonHandler(DestroyOptions, db, config)).Methods("OPTIONS")
	m.HandleFunc("/destroy/{projectId:[0-9]+}/{type}", MakeJsonHandler(Destroy, db, config)).Methods("DELETE")
	// Update
	m.HandleFunc("/update/{projectId:[0-9]+}/{type}", MakeJsonHandler(Update, db, config)).Methods("PUT")
	m.HandleFunc("/update/{projectId:[0-9]+}/{type}", MakeJsonHandler(Update, db, config)).Methods("POST")
	m.HandleFunc("/update/{projectId:[0-9]+}/{type}", MakeJsonHandler(OptionsUpdate, db, config)).Methods("OPTIONS")

	// Save
	m.HandleFunc("/save/{projectId:[0-9]+}", MakeJsonHandler(Save, db, config)).Methods("POST")
	m.HandleFunc("/save/{projectId:[0-9]+}", MakeJsonHandler(SaveOptions, db, config)).Methods("OPTIONS")
	m.HandleFunc("/saveBackup/{projectId:[0-9]+}", MakeJsonHandler(SaveBackup, db, config)).Methods("POST")
	m.HandleFunc("/saveBackup/{projectId:[0-9]+}", MakeJsonHandler(SaveOptions, db, config)).Methods("OPTIONS")
	m.HandleFunc("/saveSettings/{projectId:[0-9]+}", MakeJsonHandler(SaveSettings, db, config)).Methods("POST")
	m.HandleFunc("/saveSettings/{projectId:[0-9]+}", MakeJsonHandler(SaveOptions, db, config)).Methods("OPTIONS")

	tmpl["index.html"] = template.Must(template.ParseFiles("static/templates/watch/index.html", "static/templates/watch/base.html"))
	return m
}

func MakeJsonHandler(fn func(http.ResponseWriter, *http.Request, *sql.DB, *ini.Dict) (int, interface{}), db *sql.DB, config *ini.Dict) http.HandlerFunc {
	// do some checking, e.g. db - alpha or beta
	return func(w http.ResponseWriter, r *http.Request) {
		// for testing
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		responseCode, response := fn(w, r, db, config)
		if response != nil {
			w.WriteHeader(responseCode)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				panic(err)
			}
		}
	}
}
