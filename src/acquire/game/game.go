package game

import (
	//"acquire/common"
	"fmt"
	"html/template"
	//"log"
	"net/http"
	//"reflect"
)

// TODO: uncomment this for production
//var tmpl = template.Must(template.ParseFiles("static/templates/user/dashboard/index.html", "static/templates/user/base.html"))

func newgame(w http.ResponseWriter, r *http.Request) {
	fmt.Println("game::newgame")

	// for now parse every request so I don't have to recompile
	tmpl := template.Must(template.ParseFiles("static/templates/game/newgame.html", "static/templates/game/base.html"))

	wsURL := fmt.Sprintf("ws://%s/ws/", r.Host)
	tmpl.ExecuteTemplate(w, "base", wsURL)
}
