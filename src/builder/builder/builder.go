package builder

import (
	"builder/ini"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type JsonResponse struct {
	Message string
}

type Node struct {
	Id             string  `json: "id"`
	ParentId       string  `json: "parentId"`
	SortIndex      string  `json: "index"`
	Text           string  `json: "text"`
	Title          string  `json: "title"`
	Mtype          string  `json: "mType"`
	ShowTopicPages bool    `json: "showTopicPages"`
	Leaf           bool    `json: "leaf"`
	Children       []*Node `json:"page,omitempty"`
}

func (this *Node) Size() int {
	var size int = len(this.Children)
	for _, c := range this.Children {
		size += c.Size()
	}
	return size
}

func (this *Node) Add(nodes ...*Node) bool {
	var size = this.Size()
	for _, n := range nodes {
		if n.ParentId == this.Id {
			this.Children = append(this.Children, n)
		} else {
			for _, c := range this.Children {
				if c.Add(n) {
					break
				}
			}
		}
	}
	return this.Size() == size+len(nodes)
}

/*func (m *Node) Scan(src interface{}) error {
	srcArray := src.([]uint8)
	strValue := make([]byte, len(srcArray))
	for i, v := range srcArray {
		if v < 0 {
			strValue[i] = byte(256 + int(v))
		} else {
			strValue[i] = byte(v)
		}
	}

	return json.Unmarshal([]byte(strValue), m)
}*/

type NodeRow struct {
	Id         string         `db:"id"`
	RecordId   string         `db:"recordId"`
	Title      sql.NullString `db:"title"`
	ProjectId  string         `db:"projectId"`
	Leaf       bool           `db:"leaf"`
	ParentId   sql.NullString `db:"parentId"`
	SortIndex  int            `db:"sortIndex"`
	NodeData   string         `db:"data"`
	UpdateTime string         `db:"updateTime"`
	CreateTime string         `db:"createTime"`
	//NodeData   Node           `db:"data"`
}

var idRegexp = regexp.MustCompile("^(Topic|Page)Model-(\\d+)$")

func CreateNodeRowFromJson(jsonMap map[string]interface{}) (NodeRow, error) {
	nodwRow := NodeRow{}
	id := jsonMap["id"].(string)
	if !idRegexp.MatchString(id) && id != "root" {
		return nodwRow, errors.New(fmt.Sprintf("Invalid Id%s", id))
	}
	nodwRow.RecordId = id
	// title
	var titleNS sql.NullString
	if err := titleNS.Scan(jsonMap["title"]); err != nil {
		return nodwRow, errors.New(fmt.Sprintf("Invalid Title%s", jsonMap["title"]))
	}
	nodwRow.Title = titleNS
	// leaf
	nodwRow.Leaf = jsonMap["leaf"].(bool)
	// parentId
	var parentIdNS sql.NullString
	if err := parentIdNS.Scan(jsonMap["parentId"]); err != nil {
	}
	nodwRow.ParentId = parentIdNS
	// index
	index := int(jsonMap["index"].(float64))
	if index < 0 {
		return nodwRow, errors.New(fmt.Sprintf("Invalid index:%s", index))
	}
	// data
	jsonString, _ := json.Marshal(jsonMap)
	jsonString = bytes.Replace(jsonString, []byte("\\u003c"), []byte("<"), -1)
	jsonString = bytes.Replace(jsonString, []byte("\\u003e"), []byte(">"), -1)
	jsonString = bytes.Replace(jsonString, []byte("\\u0026"), []byte("&"), -1)
	nodwRow.NodeData = string(jsonString)
	return nodwRow, nil
}

func Read(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	treeType := vars["type"]
	fmt.Println("Tree Type::", treeType)
	if treeType != "sco" && treeType != "glossary" {
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(JsonResponse{"Invalid Input: 'type'"}); err != nil {
			panic(err)
		}
		return
	}

	nodes := GetNodeData("1", "root", db)
	fmt.Println("[" + strings.Join(nodes, ",") + "]")

	//
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[" + strings.Join(nodes, ",") + "]"))
}

func GetNodeData(projectId string, parentId string, db *sql.DB) []string {
	rows, err := db.Query("SELECT * FROM ProjectData WHERE projectId = ? AND parentId = ? ORDER BY sortIndex ASC", projectId, parentId)
	if err != nil {
		log.Fatal(err)
	}
	nodes := make([]string, 0)
	defer rows.Close()
	for rows.Next() {
		row := NodeRow{}
		if err := rows.Scan(&row.Id, &row.RecordId, &row.Title, &row.ProjectId, &row.Leaf, &row.ParentId, &row.SortIndex, &row.NodeData, &row.UpdateTime, &row.CreateTime); err != nil {
			log.Fatal(err)
		}
		nodes = append(nodes, row.NodeData)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return nodes
}

// TODO
func TestCreate(projectId string, db *sql.DB) {
	var newTestString = []byte(`{
		"recordId": "PageModel-3",
		"imgPos": "left",
		"imageWidth": "40",
		"autoPlayMedia": false,
		"nType": null,
		"title": "NEW TITLE...............",
		"linkID": "PageModel-3",
		"pText": {
			"#text": " - New Page - Text Page"
		},
		"id": "PageModel-3"
	}`)
	newNode := map[string]interface{}{}
	if err := json.Unmarshal(newTestString, &newNode); err != nil {
		panic(err)
	}

	// insert and update
	if currentNode, err := getNodeById(newNode["id"].(string), projectId, db); err != nil {
		// create
		fmt.Println("Create")
		nodeRow, _ := CreateNodeRowFromJson(newNode)
		nodeRow.ProjectId = projectId;
		TODO: call insert
	} else {
		// update
		oldNode := map[string]interface{}{}
		if err := json.Unmarshal([]byte(currentNode.NodeData), &oldNode); err != nil {
			panic(err)
		}
		MergeNodes(oldNode, newNode)
		nodeRow, _ := CreateNodeRowFromJson(oldNode)
		nodeRow.ProjectId = projectId;
		TODO: call Update
	}
}

func MergeNodes(oldNode map[string]interface{}, newNode map[string]interface{}) {
	for key, value := range newNode {
		oldNode[key] = value
	}
}

func Create(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}
func InsertNode(node NodeRow, db *sql.DB) (sql.Result, error) {
	return db.Exec("INSERT INTO ProjectData (recordId, title, projectId, leaf, parentId, sortIndex, data, updateTime, createTime) VALUES (?,?,?,?,?,?,?,NOW(),NOW())", node.RecordId, node.Title, node.ProjectId, node.Leaf, node.ParentId, node.SortIndex, node.NodeData)
}

// Delete
func Destroy(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}
func Update(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}
func Save(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {
}
func ExportJson(projectId string, treeType string, db *sql.DB) {
	if treeType != "sco" && treeType != "glossary" {
		panic("invalid type")
	}
	// get root
	rootNode := map[string]interface{}{}
	var defaultRoot = []byte(`{"complete":false,"expanded":true,"id":"root","index":0,"leaf":false,"pText":"","pType":"","parentId":null,"text":"Root","title":"root"}`)
	if rootRow, err := getNodeById("root", projectId, db); err == nil {
		defaultRoot = []byte(rootRow.NodeData)
	}
	if err := json.Unmarshal(defaultRoot, &rootNode); err != nil {
		panic(err)
	}

	addChildNodes(rootNode, projectId, db)

	json, _ := json.Marshal(rootNode)
	json = bytes.Replace(json, []byte("\\u003c"), []byte("<"), -1)
	json = bytes.Replace(json, []byte("\\u003e"), []byte(">"), -1)
	json = bytes.Replace(json, []byte("\\u0026"), []byte("&"), -1)
	if err := ioutil.WriteFile(fmt.Sprintf("C:\\data\\www\\sandbox\\ExtBuilder\\tools\\project_%s_%s.json", projectId, "sco___"), json, 0644); err != nil {
		panic(err)
	}
}

func addChildNodes(parentNode map[string]interface{}, projectId string, db *sql.DB) {
	rows, err := db.Query("SELECT * FROM ProjectData WHERE projectId = ? AND parentId = ? ORDER BY sortIndex ASC", projectId, parentNode["id"])
	if err != nil {
		log.Fatal(err)
	}
	nodes := make([]map[string]interface{}, 0)
	defer rows.Close()
	for rows.Next() {
		row := NodeRow{}
		if err := rows.Scan(&row.Id, &row.RecordId, &row.Title, &row.ProjectId, &row.Leaf, &row.ParentId, &row.SortIndex, &row.NodeData, &row.UpdateTime, &row.CreateTime); err != nil {
			log.Fatal(err)
		}
		// convert data to json map
		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(row.NodeData), &m); err != nil {
			panic(err)
		}
		nodes = append(nodes, m)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if len(nodes) > 0 {
		for _, node := range nodes {
			addChildNodes(node, projectId, db)
		}
		parentNode["page"] = nodes
	}
}

func getNodeById(recordId string, projectId string, db *sql.DB) (NodeRow, error) {
	row := NodeRow{}
	err := db.QueryRow("SELECT * FROM ProjectData WHERE projectId = ? AND recordId = ? ORDER BY sortIndex ASC", projectId, recordId).Scan(&row.Id, &row.RecordId, &row.Title, &row.ProjectId, &row.Leaf, &row.ParentId, &row.SortIndex, &row.NodeData, &row.UpdateTime, &row.CreateTime)
	switch {
	case err == sql.ErrNoRows:
		return row, errors.New(fmt.Sprintf("No Row Found with recordId:%s, projectId:%s", recordId, projectId))
	}
	return row, err
}

func SaveBackup(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}
func SaveSettings(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}

// Import
func Import(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}

func ImportJson(projectId string, treeType string, db *sql.DB) error {
	if treeType != "sco" && treeType != "glossary" {
		return errors.New("Can't find user")
	}

	// get file
	filePath := fmt.Sprintf("C:\\data\\www\\sandbox\\ExtBuilder\\tools\\project_%s_%s.json", projectId, treeType)
	file, e := ioutil.ReadFile(filePath)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	// Creating the maps for JSON
	m := map[string]interface{}{}
	if err := json.Unmarshal(file, &m); err != nil {
		panic(err)
	}
	// Flatten map
	topicCounter = 1
	pageCounter = 1
	var parentId sql.NullString
	nodes := parseMap(m, parentId, 0)
	if len(nodes) == 0 {
		panic("no nodes")
	}

	// delete old nodes
	if _, err := db.Exec("DELETE FROM ProjectData WHERE projectId = ?", projectId); err != nil {
		panic(err)
	}

	// insert new nodes
	for _, node := range nodes {
		_, err := db.Exec("INSERT INTO ProjectData (recordId, title, projectId, leaf, parentId, sortIndex, data, updateTime, createTime) VALUES (?,?,?,?,?,?,?,NOW(),NOW())", node.RecordId, node.Title, projectId, node.Leaf, node.ParentId, node.SortIndex, node.NodeData)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

var topicCounter int
var pageCounter int

func parseMap(aMap map[string]interface{}, parentId sql.NullString, sortIndex int) []NodeRow {
	nodes := []NodeRow{}
	var title sql.NullString
	if err := title.Scan(aMap["title"]); err != nil {
		panic(err)
	}

	var nodeRow NodeRow
	recordId := ""
	if _, hasChildren := aMap["page"]; hasChildren {
		if id, _ := aMap["id"]; id == "root" {
			recordId = "root"
		} else {
			recordId = "TopicModel-" + strconv.Itoa(topicCounter)
			topicCounter += 1
		}
		// Topic
		nodeRow = NodeRow{
			RecordId:  recordId,
			Title:     title,
			Id:        recordId,
			ParentId:  parentId,
			Leaf:      false,
			SortIndex: sortIndex,
		}
	} else {
		recordId = "PageModel-" + strconv.Itoa(pageCounter)
		pageCounter += 1
		// Page
		nodeRow = NodeRow{
			RecordId:  recordId,
			Title:     title,
			Id:        recordId,
			ParentId:  parentId,
			Leaf:      true,
			SortIndex: sortIndex,
		}
	}
	var childNodes []NodeRow
	// copy node data
	nodeData := make(map[string]interface{})
	for key, value := range aMap {
		switch value.(type) {
		case []interface{}:
			// page
			var pid sql.NullString
			if err := pid.Scan(recordId); err != nil {
				panic(err)
			}
			childNodes = parseArray(value.([]interface{}), pid)
		default:
			// make sure id and parentId are in sync with data
			if key == "id" {
				nodeData["id"] = recordId
			} else if key == "parentId" {
				p, _ := parentId.Value()
				nodeData["parentId"] = p
			} else {
				// copy everything else as is
				nodeData[key] = value
			}
		}
	}
	// convert node to json string
	b, err := json.Marshal(nodeData)
	if err != nil {
		fmt.Println("error:", err)
	}
	b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
	b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
	b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	nodeRow.NodeData = string(b)

	// append
	nodes = append(nodes, nodeRow)
	nodes = append(nodes, childNodes...)
	return nodes
}

func parseArray(anArray []interface{}, parentId sql.NullString) []NodeRow {
	nodes := []NodeRow{}
	for i, val := range anArray {
		switch val.(type) {
		case map[string]interface{}:
			// List of pages
			childNodes := parseMap(val.(map[string]interface{}), parentId, i)
			nodes = append(nodes, childNodes...)
		case []interface{}:
			// not used???
			parseArray(val.([]interface{}), parentId)
		default:
			// not used???
		}
	}
	return nodes
}
