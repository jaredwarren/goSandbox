package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"

	//"litmosauthor.com/unison/common"
	"litmosauthor.com/unison/ini"
	"litmosauthor.com/unison/project"
	"litmosauthor.com/unison/user"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

type Page struct {
	Title string
	Body  []byte
}

var router = mux.NewRouter()

var alphaDB *sql.DB

var config ini.Dict
var err error

//var store = sessions.NewCookieStore([]byte("something-very-secret"))

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
	// TODO: select alpha or beta db based on cust id/subdomain, need a list of db refs

	// Routs
	r := router
	//r.HandleFunc("/static/{path:.*}", common.StaticHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/project/", project.MakeMuxer("/project/", alphaDB))
	http.Handle("/user/", user.MakeMuxer("/user/", alphaDB))

	// wait for clients
	http.Handle("/", r)
	fmt.Println("Running...\n")
	http.ListenAndServe(":8080", nil)
}
