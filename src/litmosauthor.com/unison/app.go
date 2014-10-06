package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"

	//"litmosauthor.com/unison/common"
	//"litmosauthor.com/unison/bcrypt"
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

	// hash and verify a password with random salt
	/*

		//password2 := "$5$rounds=100000$fd37fa1308ad6d1d$VBIe6L0E2keL3Ne4KS0z44/44qina3HekS6pmWn5R5C"
		//password2 := []byte("fd37fa1308ad6d1d$VBIe6L0E2keL3Ne4KS0z44/44qina3HekS6pmWn5R5C")
		//data := []byte("admin1") // const Size = 32 bytes
		//password := []byte("admin1")
		password := "admin1"

		TODO: fiture out how to hash like php crypt

		salt, _ := bcrypt.Salt(10000)
		fmt.Printf("Salt: %s\n", salt)
		hash, _ := bcrypt.Hash(password, salt)

		if bcrypt.Match(password, hash) {
			fmt.Println("They match")
		} else {
			fmt.Println("NO match")
		}

		fmt.Printf("SHA256 checksum : %v\n", hash)
	*/

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
