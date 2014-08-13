package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)


type Page struct {
	Title string
	// make rune to support unicode
	Body []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}


func viewHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: strip ending slash...
	title := r.URL.Path[len("/view/"):]
	//fmt.Fprintf(w, "test:%s", title)
	p, _ := loadPage(title)
	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}


func main() {
    http.HandleFunc("/view/", viewHandler)
	http.ListenAndServe(":8080", nil)
}