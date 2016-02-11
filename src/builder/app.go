package main

import (
	"builder/builder"
	"builder/ini"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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

	builder.TestCreate("0", alphaDB)

	// router
	r := router
	//
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	//
	http.Handle("/builder/", builder.MakeMuxer("/builder/", alphaDB, &config))
	//http.Handle("/channels/", channel.MakeMuxer("/channels/", alphaDB, &config))
	http.Handle("/", r)
	fmt.Println("Running... :8081\n")
	//http.ListenAndServe(":8081", nil)
}
