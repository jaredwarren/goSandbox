package watch

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
	"youtube/channel"
	"youtube/ini"
)

type Page struct {
	Channels channel.ChannelList
	Title    string
	Videos   []channel.Video
}

func All(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	vars := mux.Vars(r)
	tag := vars["tag"]
	fmt.Println("Watch::", tag)

	channelList := channel.GetChannelList(db, tag)
	channelList = channel.GetChannelVideos(channelList)

	allVideos := make([]channel.Video, 0)
	for _, channel := range channelList.Channels {
		allVideos = append(allVideos, channel.Videos...)
	}

	// sort
	sort.Sort(channel.ByTime(allVideos))

	pagedata := &Page{
		Title:    tag,
		Channels: channelList,
		Videos:   allVideos,
	}

	tmpl["index.html"].ExecuteTemplate(w, "base", pagedata)
}
