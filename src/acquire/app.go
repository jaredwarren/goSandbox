package main

import (
	"acquire/conn"
	_ "acquire/game"
	"acquire/ini"
	"acquire/user"
	"database/sql"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

var router = mux.NewRouter()

var alphaDB *sql.DB
var config ini.Dict
var err error

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ProductsHandler::Dashboard")
}

func main() {
	config, err = ini.Load("ini/config.ini")
	if err != nil {
		log.Fatal("Failed to load config")
	}

	// setup DB
	dbName, found := config.GetString("alphadb", "name")
	if !found {
		log.Fatal("Couldn't get name")
	}
	dbUser, found := config.GetString("alphadb", "user")
	if !found {
		log.Fatal("Couldn't get user")
	}
	dbPassword, found := config.GetString("alphadb", "password")
	if !found {
		log.Fatal("Couldn't get password")
	}
	alphaDB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", dbUser, dbPassword, dbName))
	if err != nil {
		log.Fatal(err)
	}

	// Routs
	r := router
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	//http.Handle("/game/", game.MakeMuxer("/game/", alphaDB, config))
	http.Handle("/user/", user.MakeMuxer("/user/", alphaDB, &config))

	// websocket
	http.Handle("/ws/", conn.MakeMuxer("/ws/", alphaDB, &config))

	// wait for clients
	http.Handle("/", r)
	fmt.Println("Running... :8080\n")
	http.ListenAndServe(":8080", nil)
}
