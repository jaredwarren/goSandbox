package main

import (
	//"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"regexp"
	"flag"
	"net"
	"log"
	"io/ioutil"
)
var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

// handlerFunc adapts a function to an http.Handler.
// type handlerFunc func(http.ResponseWriter, *http.Request) error

// func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	if r.Host != "www.gorillatoolkit.org" && !appengine.IsDevAppServer() {
// 		r.URL.Host = "www.gorillatoolkit.org"
// 		http.Redirect(w, r, r.URL.String(), 301)
// 		return
// 	}

// 	err := f(w, r)
// 	if err != nil {
// 		appengine.NewContext(r).Errorf("Error %s", err.Error())
// 		if e, ok := err.(doc.GetError); ok {
// 			http.Error(w, "Error getting files from "+e.Host+".", http.StatusInternalServerError)
// 		} else if appengine.IsCapabilityDisabled(err) || appengine.IsOverQuota(err) {
// 			http.Error(w, "Internal error: "+err.Error(), http.StatusInternalServerError)
// 		} else {
// 			http.Error(w, "Internal Error", http.StatusInternalServerError)
// 		}
// 	}
// }

type Page struct {
	Title string
	Body  []byte
}

var validPath = regexp.MustCompile("^/(edit|save|view|home)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplate(w, "404", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := &Page{Title: "title", Body: []byte("body")}
	executeTemplate(w, "home", p)
}

var templates = template.Must(template.ParseFiles("static/templates/404.html", "static/templates/home.html"))
func executeTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var router = mux.NewRouter()

func main() {
	flag.Parse()
	http.HandleFunc("/view/", makeHandler(homeHandler))
	if *addr {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
		if err != nil {
			log.Fatal(err)
		}
		s := &http.Server{}
		s.Serve(l)
		return
	}

	http.ListenAndServe(":8080", nil)

	//r := router
	//r.Handle("/", makeHandler(homeHandler))
	//http.Handle("/", r)
	//r.Handle("/dashboard" handlerFunc(dashboardHandler))

}
