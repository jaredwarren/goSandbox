package conn

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"net/http"
)

func MakeMuxer(prefix string, db *sql.DB) http.Handler {
	var m *mux.Router
	if prefix == "" {
		m = mux.NewRouter()
	} else {
		m = mux.NewRouter().PathPrefix(prefix).Subrouter()
	}

	// start hub
	h := newHub()
	go h.run()

	// ws routes
	var wr *SocketRouter
	wr = NewStockRouter()
	wr.HandleFunc("message", messageHandler)
	wr.HandleFunc("message2", messageHandler2)

	// catch everything here!!!!
	m.Handle("/", wsHandler{h: h, wr: wr})

	return m
}

// just here to test, should move somewhere else
func messageHandler(message msg) {
	fmt.Println("Action:", message.Action, " Data:", message.Message)
}

func messageHandler2(message msg) {
	fmt.Println("Action2:", message.Action, " Data:", message.Message)
}
