package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	//"html/template"
	"net/http"
	"regexp"
	//"net"
	"log"
	//"io/ioutil"

	"litmosauthor.com/unison/project"

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

var validPath = regexp.MustCompile("^/(edit|save|view|home)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, *sql.DB)) http.HandlerFunc {
	// do some checking, e.g. db - alpha or beta
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, alphaDB)
	}
}

/*
func notFoundHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	executeTemplate(w, "404", nil)
	db, _ := sql.Open("mysql", "user:password@/dbname")
	age := 27
    rows, err := db.Query("SELECT name FROM users WHERE age=?", age)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
                log.Fatal(err)
        }
        fmt.Printf("%s is %d\n", name, age)
    }
    if err := rows.Err(); err != nil {
        log.Fatal(err)
    }
}
*/

/*func homeHandler(res http.ResponseWriter, req *http.Request, db *sql.DB) {
	fmt.Println("home")
	cust_id := "unison"
	rows, err := db.Query("SELECT project_name FROM project WHERE cust_id=?", cust_id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("--%s--\n", name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	p := &Page{Title: "title", Body: []byte("body")}
	executeTemplate(res, "home", p)
}

var templates = template.Must(template.ParseFiles("static/templates/404.html", "static/templates/home.html"))

func executeTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	fmt.Printf("Template:" + tmpl)
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}*/

// http://www.gorillatoolkit.org/pkg/mux
var router = mux.NewRouter()

var alphaDB *sql.DB

//var store = sessions.NewCookieStore([]byte("something-very-secret"))

func main() {
	r := router

	// setup DB
	var err error
	alphaDB, err = sql.Open("mysql", "webuser:(^#F$nt45T!c.?-)@/alpha")
	if err != nil {
		log.Fatal(err)
	}

	// Routs
	//r.HandleFunc("/", makeHandler(project.Dashboard))
	http.Handle("/project/", project.MakeMuxer("/project/"))
	//r.HandleFunc("/project/", project.makeHandler("/project/"))

	fmt.Println("Started...")
	// wait for clients
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
	//http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
}
