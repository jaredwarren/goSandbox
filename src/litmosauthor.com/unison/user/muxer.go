package user

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"litmosauthor.com/unison/common"
	"net/http"
)

func MakeMuxer(prefix string, db *sql.DB) http.Handler {
	var m *mux.Router
	if prefix == "" {
		m = mux.NewRouter()
	} else {
		m = mux.NewRouter().PathPrefix(prefix).Subrouter()
	}
	m.HandleFunc("/login/", common.MakeHandler(loginForm, db)).Methods("GET")
	m.HandleFunc("/login/", common.MakeHandler(login, db)).Methods("POST")
	//m.HandleFunc("/{path:.*}", common.MakeHandler(common.NotFoundFunc, db))
	m.HandleFunc("/logout/", common.MakeHandler(logout, db)).Methods("GET")
	m.HandleFunc("/dashboard/", common.MakeHandler(dashboard, db)).Methods("GET")
	return m
}
