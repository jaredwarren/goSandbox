package watch

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"sort"
	"time"
	"youtube/channel"
	"youtube/ini"
)

type Page struct {
	Channels channel.ChannelList
	Title    string
	Videos   []channel.Video
}

// sort
type ByTime []channel.Video

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].Time.Before(a[j].Time) }

// Time
const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * 60
	secondsPerDay    = 24 * secondsPerHour
	secondsPerWeek   = 7 * secondsPerDay
	secondsPerMonth  = 30 * secondsPerDay
	secondsPerYear   = 52 * secondsPerWeek
)

type TimeMatch struct {
	Reg    *regexp.Regexp
	Offset time.Time
}

func Test(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	fmt.Println("Watch::")

}

func All(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	vars := mux.Vars(r)
	title := vars["tag"]
	fmt.Println("Watch::", title)

	channelList := channel.GetChannelList(db)
	channelList = channel.GetChannelVideos(channelList)

	allVideos := make([]channel.Video, 0)
	for _, channel := range channelList.Channels {
		allVideos = append(allVideos, channel.Videos...)
	}

	sort.Sort(ByTime(allVideos))

	pagedata := &Page{
		Title:    title,
		Channels: channelList,
		Videos:   allVideos,
	}

	tmpl["index.html"].ExecuteTemplate(w, "base", pagedata)
}

//var notRe = regexp.MustCompile("^not(.+?)$")
