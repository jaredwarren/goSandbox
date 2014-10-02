package project

import (
	"database/sql"
	//"fmt"
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
	m.HandleFunc("/", common.MakeHandler(Dashboard, db))
	//m.HandleFunc("/{path:.*}", common.MakeHandler(common.NotFoundFunc, db))
	return m
}
