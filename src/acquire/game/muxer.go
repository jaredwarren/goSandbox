package game

import (
	//"fmt"
	"acquire/common"
	"github.com/gorilla/mux"
	"net/http"
)

func MakeMuxer(prefix string) http.Handler {
	var m *mux.Router
	if prefix == "" {
		m = mux.NewRouter()
	} else {
		m = mux.NewRouter().PathPrefix(prefix).Subrouter()
	}

	m.HandleFunc("/newgame/", common.MakeHandler(newgame)).Methods("GET")
	return m
}
