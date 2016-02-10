package tag

import (
	"database/sql"
	//"fmt"
	//"github.com/gorilla/mux"
	"log"
	"net/http"
	"youtube/ini"
)

type Page struct {
	Channels map[string]map[string]bool
	Title    string
}

type Tag struct {
	TagId   string
	TagName string
}

func ListTags(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

	/*channels := GetChannelList(db)

	pagedata := &Page{
		Title:    "Channels",
		Channels: channels,
	}

	tmpl["list.html"].ExecuteTemplate(w, "base", pagedata)*/
}

type TagListRow struct {
	TagId   string `db:"id"`
	TagName string `db:"name"`
}

func GetTagList(db *sql.DB) []Tag {
	var tagList = []Tag{}

	rows, err := db.Query("SELECT tag.* FROM tag")
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		//var project Project
		row := TagListRow{}
		if err := rows.Scan(&row.TagId, &row.TagName); err != nil {
			log.Fatal(err)
		}
		tagList = append(tagList, Tag{TagId: row.TagId, TagName: row.TagName})
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return tagList
}

/*

type ChannelListRow struct {
	Url  string `db:"url"`
	Name string `db:"name"`
}

func GetChannelList(db *sql.DB) map[string]map[string]bool {
	var channelList = make(map[string]map[string]bool)

	rows, err := db.Query("SELECT  channels.url, tag.name FROM channels LEFT JOIN channel_x_tag ON channels.id = channel_x_tag.channelId LEFT JOIN tag ON channel_x_tag.tagId = tag.id")
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		//var project Project
		row := ChannelListRow{}
		if err := rows.Scan(&row.Url, &row.Name); err != nil {
			log.Fatal(err)
		}
		if _, ok := channelList[row.Url]; !ok {
			channelList[row.Url] = make(map[string]bool)
		}
		channelList[row.Url][row.Name] = true
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return channelList
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
*/
