package watch

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"
	"youtube/ini"
)

// request
type SPF struct {
	Title string   `json: "title"`
	Head  string   `json: "head"`
	Body  SPF_Body `json: "body"`
	Foot  string   `json: "foot"`
}

type SPF_Body struct {
	Content string `json: "content"`
}

// response
type Page struct {
	Channels []*Channel
	Title    string
	Videos   []Video
}

type Channel struct {
	Title       string
	ChannelIcon string
	Videos      []Video
}
type Video struct {
	Title       string
	TimeString  string
	ChannelName string
	ChannelId   string
	ChannelIcon string
	Time        time.Time
	Id          string
}

// sort
type ByTime []Video

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

var regSec = regexp.MustCompile("(\\d+?) seconds?")
var regMin = regexp.MustCompile("(\\d+?) minutes?")
var regHor = regexp.MustCompile("(\\d+?) hours?")
var regDay = regexp.MustCompile("(\\d+?) days?")
var regWek = regexp.MustCompile("(\\d+?) weeks?")
var regMon = regexp.MustCompile("(\\d+?) months?")
var regYer = regexp.MustCompile("(\\d+?) years?")

type TimeMatch struct {
	Reg    *regexp.Regexp
	Offset time.Time
}

func strtotime(timeString string) time.Time {
	t1 := time.Now()
	switch {
	case regSec.MatchString(timeString):
		timeMatch := regSec.FindAllStringSubmatch(timeString, -1)[0][1]
		i, _ := strconv.Atoi(timeMatch)
		t1 = time.Date(0, 0, 0, 0, 0, i, 0, time.UTC)
	case regMin.MatchString(timeString):
		timeMatch := regMin.FindAllStringSubmatch(timeString, -1)[0][1]
		i, _ := strconv.Atoi(timeMatch)
		t1 = time.Date(0, 0, 0, 0, i, 0, 0, time.UTC)
	case regHor.MatchString(timeString):
		timeMatch := regHor.FindAllStringSubmatch(timeString, -1)[0][1]
		i, _ := strconv.Atoi(timeMatch)
		t1 = time.Date(0, 0, 0, i, 0, 0, 0, time.UTC)
	case regDay.MatchString(timeString):
		timeMatch := regDay.FindAllStringSubmatch(timeString, -1)[0][1]
		i, _ := strconv.Atoi(timeMatch)
		t1 = time.Date(0, 0, i, 0, 0, 0, 0, time.UTC)
	case regWek.MatchString(timeString):
		timeMatch := regWek.FindAllStringSubmatch(timeString, -1)[0][1]
		i, _ := strconv.Atoi(timeMatch)
		t1 = time.Date(0, 0, i*7, 0, 0, 0, 0, time.UTC)
	case regMon.MatchString(timeString):
		timeMatch := regMon.FindAllStringSubmatch(timeString, -1)[0][1]
		i, _ := strconv.Atoi(timeMatch)
		t1 = time.Date(0, time.Month(i), 0, 0, 0, 0, 0, time.UTC)
	case regYer.MatchString(timeString):
		timeMatch := regYer.FindAllStringSubmatch(timeString, -1)[0][1]
		i, _ := strconv.Atoi(timeMatch)
		t1 = time.Date(i, 0, 0, 0, 0, 0, 0, time.UTC)
	default:
		fmt.Println("It doesn't match", timeString)
	}
	return t1
}

// stuff
func getChannelVideos(channelUrl string) *Channel {
	resp, err := http.Get("https://www.youtube.com/" + channelUrl + "/videos?spf=navigate")
	if err != nil {
		fmt.Println(".....")
		log.Fatal(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data SPF
	err = decoder.Decode(&data)

	bodyString := string(data.Body.Content)
	titleRe := regexp.MustCompile("\"channel-header-profile-image\".src=\"(.+?)\".+?title=\"(.+?)\"")
	matches := titleRe.FindAllStringSubmatch(bodyString, -1)[0]
	channelIcon := matches[1]
	title := matches[2]

	timeRe := regexp.MustCompile("yt-lockup-meta-info.+?<li>.+?<li>(.+?)<")
	timeMatches := timeRe.FindAllStringSubmatch(bodyString, -1)

	idRe := regexp.MustCompile("/watch\\?v=(.+?)\">(.+?)<")
	idMatches := idRe.FindAllStringSubmatch(bodyString, -1)

	videos := make([]Video, len(timeMatches))

	for i := 0; i < len(idMatches); i++ {
		videos[i] = Video{
			Id:          idMatches[i][1],
			ChannelName: title,
			ChannelId:   channelUrl,
			ChannelIcon: channelIcon,
			Title:       html.UnescapeString(idMatches[i][2]),
			Time:        strtotime(timeMatches[i][1]),
			TimeString:  timeMatches[i][1],
		}
	}
	return &Channel{title, channelIcon, videos}
}

func getAsyncChannelVideos(channels []string) ([]*Channel, int) {
	ch := make(chan *Channel, len(channels))
	responses := []*Channel{}
	totalVideos := 0
	for _, url := range channels {
		go func(url string) {
			resp := getChannelVideos(url)
			ch <- resp
		}(url)
	}

	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			totalVideos += len(r.Videos)
			if len(responses) == len(channels) {
				return responses, totalVideos
			}
		case <-time.After(2000 * time.Millisecond):
			fmt.Printf("URL timout")
		}
	}

	return responses, totalVideos
}

var channelMaps = map[string]map[string]bool{
	"user/SargonofAkkad100":            {"other": true},
	"channel/UC6cMYsKMx6XicFcFm7mTsmA": {"other": true}, // sargon live
	"channel/UCMIj-wEiKIcGAcLoBO2ciQQ": {"other": true}, // tl:dr
	"channel/UClzNJ7y2Q6wY0tEOzE6EM9Q": {"other": true}, // Vernaculis
	"channel/UCC1rjUKeELaSKsxg0O1bNGw": {"other": true}, // harmful opnions
	"user/noelplum99":                  {"other": true},
	"user/armouredskeptic":             {"other": true},
	"user/SecularTalk": {
		"news":  true,
		"other": true,
	},
	"user/MidweekPolitics": {
		"news":  true,
		"other": true,
	},
	"user/MundaneMatt": {
		"news":  true,
		"other": true,
	},
	"user/TheAmazingAtheist":           {"other": true},
	"channel/UCla6APLHX6W3FeNLc8PYuvg": {"other": true}, // Lauren Southern
	"user/GEdwardsPhilosophy":          {"other": true},
	"user/bigthink":                    {"other": true},
	"user/Shoe0nHead":                  {"other": true},
	"user/yiannopoulosm":               {"other": true},
	"user/IHEOfficial":                 {"other": true},
	"user/ngoroff":                     {"other": true},
	"user/Thunderf00t":                 {"other": true},
	"channel/UCfYbb7nga6-icsFWWgS-kWw": {"other": true},
	"user/sparrowtm":                   {"other": true}, // Top Hats and Champagne
	"channel/UCTrecbx23AAYdmFHDkci0aQ": {"other": true}, // undoomed
}

func Test(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	fmt.Println("Watch::")

}

func All(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	vars := mux.Vars(r)
	title := vars["tag"]
	fmt.Println("Watch::", title)

	channelList := getChannels(title)

	c := getChannelList(db)
	fmt.Println(c)

	results, _ := getAsyncChannelVideos(channelList)
	allVideos := make([]Video, 0)
	for _, channel := range results {
		allVideos = append(allVideos, channel.Videos...)
	}

	sort.Sort(ByTime(allVideos))

	pagedata := &Page{
		Title:    title,
		Channels: results,
		Videos:   allVideos,
	}

	tmpl["index.html"].ExecuteTemplate(w, "base", pagedata)
}

type ChannelListRow struct {
	Url  string `db:"url"`
	Name string `db:"name"`
}

func getChannelList(db *sql.DB) map[string]map[string]bool {
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

var notRe = regexp.MustCompile("^not(.+?)$")

func getChannels(tag string) []string {
	var channelList []string

	if notRe.MatchString(tag) {
		matches := notRe.FindAllStringSubmatch(tag, -1)[0]
		tag = matches[1]
		for channelKey, channelMap := range channelMaps {
			if _, ok := channelMap[tag]; !ok {
				channelList = append(channelList, channelKey)
			}
		}
	} else {
		for channelKey, channelMap := range channelMaps {
			if _, ok := channelMap[tag]; ok {
				channelList = append(channelList, channelKey)
			}
		}
	}

	// defautl to all if empty
	if len(channelList) == 0 {
		for channelKey, _ := range channelMaps {
			channelList = append(channelList, channelKey)
		}
	}

	return channelList
}
