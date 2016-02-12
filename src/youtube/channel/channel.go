package channel

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"youtube/ini"
	"youtube/tag"
)

type Page struct {
	Channels []Channel
	Tags     []tag.Tag
	Title    string
}

// channel list

type ChannelList struct {
	Channels []Channel
}

func (r *ChannelList) HasChannel(channelUrl string) bool {
	for _, channel := range r.Channels {
		if channel.ChannelURL == channelUrl {
			return true
		}
	}
	return false
}
func (r *ChannelList) Size() int {
	return len(r.Channels)
}

// channel

type Channel struct {
	ChannelId   string
	ChannelURL  string
	Tags        []tag.Tag
	Title       string
	ChannelIcon string
	ChannelName string
	Videos      []Video
}

// Video

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

var regSec = regexp.MustCompile("(\\d+?) seconds?")
var regMin = regexp.MustCompile("(\\d+?) minutes?")
var regHor = regexp.MustCompile("(\\d+?) hours?")
var regDay = regexp.MustCompile("(\\d+?) days?")
var regWek = regexp.MustCompile("(\\d+?) weeks?")
var regMon = regexp.MustCompile("(\\d+?) months?")
var regYer = regexp.MustCompile("(\\d+?) years?")

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

func GetChannelVideos(channelList ChannelList) ChannelList {
	v := getAsyncChannelVideos(channelList)
	return v
}

// stuff
func attachChannelVideos(channel Channel) Channel {
	resp, err := http.Get("https://www.youtube.com/" + channel.ChannelURL + "/videos?spf=navigate")
	if err != nil {
		fmt.Println("Response Error")
		log.Fatal(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data SPF
	err = decoder.Decode(&data)

	bodyString := string(data.Body.Content)
	titleRe := regexp.MustCompile("\"channel-header-profile-image\".src=\"(.+?)\".+?title=\"(.+?)\"")
	m := titleRe.FindAllStringSubmatch(bodyString, -1)
	if len(m) < 1 {
		fmt.Println("EmptyBody...")
		fmt.Println(data)
		return channel
	}
	matches := m[0]
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
			ChannelId:   channel.ChannelURL,
			ChannelIcon: channelIcon,
			Title:       html.UnescapeString(idMatches[i][2]),
			Time:        strtotime(timeMatches[i][1]),
			TimeString:  timeMatches[i][1],
		}
	}
	channel.ChannelName = title
	channel.ChannelIcon = channelIcon
	channel.Videos = videos
	return channel
}

func getAsyncChannelVideos(channels ChannelList) ChannelList {
	ch := make(chan Channel, channels.Size())
	responses := ChannelList{}
	for _, channel := range channels.Channels {
		go func(c Channel) {
			resp := attachChannelVideos(c)
			ch <- resp
		}(channel)
	}

	for {
		select {
		case r := <-ch:
			responses.Channels = append(responses.Channels, r)
			if responses.Size() == channels.Size() {
				return responses
			}
		case <-time.After(2000 * time.Millisecond):
			fmt.Printf("URL timout")
		}
	}

	return responses
}

// API
func ListChannels(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	channels := GetChannelList(db, "all")

	pagedata := &Page{
		Title:    "Channels",
		Tags:     tag.GetTagList(db),
		Channels: channels.Channels,
	}

	tmpl["list.html"].ExecuteTemplate(w, "base", pagedata)
}

type ChannelListRow struct {
	ChannelId  string `db:"channelId"`
	ChannelUrl string `db:"channelUrl"`
	tagIds     string `db:"tagIds"`
	tagNames   string `db:"tagNames"`
}

func GetChannelList(db *sql.DB, tagFilter string) ChannelList {
	var channelList = ChannelList{}

	if tagFilter == "all" {
		tagFilter = "youtube"
	}

	rows, err := db.Query("SELECT channels.id as channelId, channels.url as channelUrl, GROUP_CONCAT(DISTINCT tag.id) as tagIds, GROUP_CONCAT(DISTINCT tag.name) as tagNames FROM channels LEFT JOIN channel_x_tag ON channels.id = channel_x_tag.channelId LEFT JOIN tag ON channel_x_tag.tagId = tag.id WHERE tag.name= ? GROUP BY channelId", tagFilter)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		row := ChannelListRow{}
		if err := rows.Scan(&row.ChannelId, &row.ChannelUrl, &row.tagIds, &row.tagNames); err != nil {
			log.Fatal(err)
		}
		tagIdList := strings.Split(row.tagIds, ",")
		tagNameList := strings.Split(row.tagNames, ",")
		var tags = make([]tag.Tag, len(tagIdList))
		for i, _ := range tagIdList {
			tags[i] = tag.Tag{tagIdList[i], tagNameList[i]}
		}

		channelList.Channels = append(channelList.Channels, Channel{ChannelId: row.ChannelId, ChannelURL: row.ChannelUrl, Tags: tags})
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return channelList
}

var channelIdRe = regexp.MustCompile("(channel|user)/(.+?)(/.*)?$")

type JsonResponse struct {
	Message string
}

func CreateChannels(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	err := r.ParseForm()
	if err != nil {
		//handle error http.Error() for example
		return
	}

	// get channel id input
	channelId := r.FormValue("channelId")
	// filter input
	if !channelIdRe.MatchString(channelId) {
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(JsonResponse{"Error..."}); err != nil {
			panic(err)
		}
		return
	}
	channelMatches := channelIdRe.FindAllStringSubmatch(channelId, -1)[0]
	channelId = channelMatches[1] + "/" + channelMatches[2]

	fmt.Printf("%+v\n", channelId)

	channelList := GetChannelList(db, "all")
	ok := channelList.HasChannel(channelId)

	if ok {
		w.WriteHeader(http.StatusConflict) // conflict
		if err := json.NewEncoder(w).Encode(JsonResponse{"Already exists"}); err != nil {
			panic(err)
		}
		return
	}
	result, err := db.Exec("INSERT INTO channels (url) VALUES (?);", channelId)
	if err != nil {
		panic(err)
	}
	newChannelId, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}

	// add tags
	tags := r.Form["tags[]"]
	if len(tags) == 0 {
		tags = append(tags, "3")
	}

	for _, tagId := range tags {
		_, err = db.Exec("INSERT INTO channel_x_tag (channelId, tagId) VALUES (?, ?);", newChannelId, tagId)
		if err != nil {
			panic(err)
		}
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(JsonResponse{"Created"}); err != nil {
		panic(err)
	}
}

func GetChannel(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	vars := mux.Vars(r)
	channelId := vars["id"]
	fmt.Println("GET::", channelId)
	w.WriteHeader(http.StatusNotImplemented) // unprocessable entity
	if err := json.NewEncoder(w).Encode(JsonResponse{"Not Implemented"}); err != nil {
		panic(err)
	}
}

func UpdateChannel(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	vars := mux.Vars(r)
	channelId := vars["id"]
	//   UPDATE `sandbox`.`tag` SET `name` = 'news' WHERE `tag`.`id` = 2;
	fmt.Println("UPDATE::", channelId)
	w.WriteHeader(http.StatusNotImplemented) // unprocessable entity
	if err := json.NewEncoder(w).Encode(JsonResponse{"Not Implemented"}); err != nil {
		panic(err)
	}
}

func DeleteChannel(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// get channel id input
	channelId := r.FormValue("channelId")
	// filter input
	if !channelIdRe.MatchString(channelId) {
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(JsonResponse{"Error..."}); err != nil {
			panic(err)
		}
		return
	}
	channelMatches := channelIdRe.FindAllStringSubmatch(channelId, -1)[0]
	channelId = channelMatches[1] + "/" + channelMatches[2]

	fmt.Printf("%+v\n", channelId)

	channelList := GetChannelList(db, "all")
	ok := channelList.HasChannel(channelId)

	if ok {
		w.WriteHeader(http.StatusConflict) // conflict
		if err := json.NewEncoder(w).Encode(JsonResponse{"Already exists"}); err != nil {
			panic(err)
		}
		return
	}
	/*result, err := db.Exec("INSERT INTO channels (url) VALUES (?);", channelId)
	if err != nil {
		panic(err)
	}
	newChannelId, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}*/
}
