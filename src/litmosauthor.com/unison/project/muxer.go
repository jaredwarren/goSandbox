package project

import (
	"fmt"
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

	m.HandleFunc("/", Dashboard)

	m.HandleFunc("/{path:.*}", NotFoundFunc)

	return m

}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Project::Dashboard")
	fmt.Printf("Dashboard")
}

func NotFoundFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Project::NotFoundFunc")
	fmt.Printf("404")
}
