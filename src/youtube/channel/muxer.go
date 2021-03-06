package channel

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

	// CRUD
	m.HandleFunc("/", common.MakeHandler(ListChannels, db, config)).Methods("GET")
	m.HandleFunc("/", common.MakeHandler(CreateChannels, db, config)).Methods("POST")
	m.HandleFunc("/", common.MakeHandler(DeleteChannel, db, config)).Methods("DELETE")
	m.HandleFunc("/", common.MakeHandler(UpdateChannel, db, config)).Methods("PUT")

	m.HandleFunc("/{id}/", common.MakeHandler(GetChannel, db, config)).Methods("GET")
	m.HandleFunc("/{id}/", common.MakeHandler(DeleteChannel, db, config)).Methods("DELETE")

	tmpl["index.html"] = template.Must(template.ParseFiles("static/templates/watch/index.html", "static/templates/watch/base.html"))
	tmpl["list.html"] = template.Must(template.ParseFiles("static/templates/channel/list.html", "static/templates/channel/base.html"))
	//tmpl["channel.html"] = template.Must(template.ParseFiles("static/templates/channel/channel.html", "static/templates/channel/base.html"))
	return m
}
