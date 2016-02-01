package channel

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"net/http"
	"time"
	"youtube/ini"
)

func ListChannels(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}

func CreateChannels(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	fmt.Println("TODO: insert channel into db")
}

func GetChannel(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	vars := mux.Vars(r)
	channelId := vars["id"]
	fmt.Println("GET::", channelId)
}

func UpdateChannel(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	vars := mux.Vars(r)
	channelId := vars["id"]
	fmt.Println("UPDATE::", channelId)
}

func DeleteChannel(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	vars := mux.Vars(r)
	channelId := vars["id"]
	fmt.Println("DELETE::", channelId)
}
