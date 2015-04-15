package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"

	"acquire/ini"

	/*"database/sql"
	_ "github.com/go-sql-driver/mysql"*/)

var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

var router = mux.NewRouter()
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

	// Routs
	r := router
	//r.HandleFunc("/static/{path:.*}", common.StaticHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/products", ProductsHandler)
	//http.Handle("/project/", project.MakeMuxer("/project/"))
	//http.Handle("/user/", user.MakeMuxer("/user/", alphaDB))

	// wait for clients
	http.Handle("/", r)
	fmt.Println("Running...\n")
	http.ListenAndServe(":8080", nil)
}