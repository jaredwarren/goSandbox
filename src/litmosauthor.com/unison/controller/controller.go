package controller



import (
	"fmt"
	//"database/sql"
    //"html/template"
    "net/http"
    //"log"
 	//_ "github.com/go-sql-driver/mysql"
)

type Controller interface {
    executeTemplate(w http.ResponseWriter, tmpl string)
}



func executeTemplate(w http.ResponseWriter, tmpl string) {
    fmt.Printf("Template:"+tmpl)
    //fmt.Printf(w)
    //err := templates.ExecuteTemplate(w, tmpl+".html", nil)
    //if err != nil {
    //    http.Error(w, err.Error(), http.StatusInternalServerError)
    //}
}