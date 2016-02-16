package builder

import (
	"builder/ini"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type JsonResponse struct {
	Message string `json: "message"`
	Success bool   `json: "success"`
}

type NodeList struct {
	Nodes map[string]interface{}
}

type Node struct {
	Id             string `json: "id"`
	ParentId       string `json: "parentId"`
	SortIndex      string `json: "index"`
	Text           string `json: "text"`
	Title          string `json: "title"`
	Mtype          string `json: "mType"`
	ShowTopicPages bool   `json: "showTopicPages"`
	Leaf           bool   `json: "leaf"`
}

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
}

var idRegexp = regexp.MustCompile("^(Topic|Page)Model-(\\d+)$")

func CreateNodeRowFromJson(jsonMap map[string]interface{}) (NodeRow, error) {
	// id
	nodeRow := NodeRow{}
	id := jsonMap["id"].(string)
	// NOTE: id regexp is only necessary for ExtJs
	if !idRegexp.MatchString(id) && id != "root" {
		return nodeRow, errors.New(fmt.Sprintf("Invalid Id%s", id))
	}
	nodeRow.RecordId = id

	// title
	var titleNS sql.NullString
	if err := titleNS.Scan(jsonMap["title"]); err != nil {
		return nodeRow, errors.New(fmt.Sprintf("Invalid Title%s", jsonMap["title"]))
	}
	nodeRow.Title = titleNS

	// leaf
	if leaf, ok := jsonMap["leaf"]; ok {
		nodeRow.Leaf = leaf.(bool)
	} else {
		// TODO: check db for other possible children, then set default
		nodeRow.Leaf = true
	}

	// parentId
	// TODO: check db for parent
	var parentIdNS sql.NullString
	parentId := "root"
	// NOTE: id regexp is only necessary for ExtJs
	if p, ok := jsonMap["parentId"]; ok && (idRegexp.MatchString(p.(string)) || p == "root") {
		parentId = p.(string)
	}
	if err := parentIdNS.Scan(parentId); err != nil {
		return nodeRow, errors.New(fmt.Sprintf("Failed to create parent%s", parentId))
	}
	nodeRow.ParentId = parentIdNS

	// index
	// TODO: check db for other possible sibblings, then set default index to last
	index := 0
	if i, ok := jsonMap["index"]; ok {
		index = int(i.(float64))
	}
	if index < 0 {
		return nodeRow, errors.New(fmt.Sprintf("Invalid index:%s", index))
	}
	nodeRow.SortIndex = index

	// data
	jsonString, _ := json.Marshal(jsonMap)
	jsonString = bytes.Replace(jsonString, []byte("\\u003c"), []byte("<"), -1)
	jsonString = bytes.Replace(jsonString, []byte("\\u003e"), []byte(">"), -1)
	jsonString = bytes.Replace(jsonString, []byte("\\u0026"), []byte("&"), -1)
	nodeRow.NodeData = string(jsonString)
	return nodeRow, nil
}

func ReadOptions(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	return http.StatusOK, JsonResponse{}
}
func Read(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	vars := mux.Vars(r)
	treeType := vars["type"]
	if treeType != "sco" && treeType != "glossary" {
		return 422, JsonResponse{"Invalid Input: 'type'", false}
	}

	nodeId := r.URL.Query().Get("node")

	nodes := GetNodeData(vars["projectId"], nodeId, db)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[" + strings.Join(nodes, ",") + "]"))
	return http.StatusOK, nil
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

var regJsonArray = regexp.MustCompile(`^\[(.|\n|\r)+\]$`)

func CreateOrUpdate(nodeData []byte, projectId string, db *sql.DB) error {
	// force Json to array
	if !regJsonArray.MatchString(string(nodeData)) {
		nodeData = append([]byte(`[`), nodeData...)
		e := []byte(`]`)
		nodeData = append(nodeData, e...)
	}
	nodes := make([]map[string]interface{}, 0)
	if err := json.Unmarshal(nodeData, &nodes); err != nil {
		return err
	}

	for _, node := range nodes {
		if currentNode, err := getNodeById(node["id"].(string), projectId, db); err != nil {
			// create
			nodeRow, _ := CreateNodeRowFromJson(node)
			nodeRow.ProjectId = projectId
			InsertNode(nodeRow, db)
		} else {
			// update
			oldNode := map[string]interface{}{}
			if err := json.Unmarshal([]byte(currentNode.NodeData), &oldNode); err != nil {
				return err
			}
			MergeNodes(oldNode, node)
			nodeRow, _ := CreateNodeRowFromJson(oldNode)
			nodeRow.ProjectId = projectId
			UpdateNode(nodeRow, db)
		}
	}

	return nil
}

func MergeNodes(oldNode map[string]interface{}, newNode map[string]interface{}) {
	for key, value := range newNode {
		oldNode[key] = value
	}
}

func CreateOptions(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	return http.StatusOK, JsonResponse{}
}
func Create(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	vars := mux.Vars(r)
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := CreateOrUpdate(body, vars["projectId"], db); err != nil {
		panic(err)
	}
	return http.StatusOK, JsonResponse{"Updated", true}
}
func InsertNode(node NodeRow, db *sql.DB) (sql.Result, error) {
	return db.Exec("INSERT INTO ProjectData (recordId, title, projectId, leaf, parentId, sortIndex, data, updateTime, createTime) VALUES (?,?,?,?,?,?,?,NOW(),NOW())", node.RecordId, node.Title, node.ProjectId, node.Leaf, node.ParentId, node.SortIndex, node.NodeData)
}
func UpdateNode(node NodeRow, db *sql.DB) (sql.Result, error) {
	return db.Exec("UPDATE ProjectData SET title = ?, leaf = ?, parentId = ?, sortIndex = ?, data = ?, updateTime = NOW() WHERE recordId = ? AND projectId = ?", node.Title, node.Leaf, node.ParentId, node.SortIndex, node.NodeData, node.RecordId, node.ProjectId)
}

// Delete
func DestroyOptions(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	return http.StatusOK, JsonResponse{}
}
func Destroy(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	vars := mux.Vars(r)
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	// force Json to array
	if !regJsonArray.MatchString(string(body)) {
		body = append([]byte(`[`), body...)
		e := []byte(`]`)
		body = append(body, e...)
	}
	nodes := make([]map[string]interface{}, 0)
	if err := json.Unmarshal(body, &nodes); err != nil {
		return http.StatusBadRequest, JsonResponse{err.Error(), false}
	}

	for _, node := range nodes {
		if _, err := db.Exec("DELETE FROM ProjectData WHERE recordId = ? AND projectId = ? LIMIT 1", node["id"], vars["projectId"]); err != nil {
			panic(err)
		}
	}

	return http.StatusOK, JsonResponse{"Updated", true}
}

func OptionsUpdate(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	// not sure what this is supposed to do
	return http.StatusOK, JsonResponse{}
}
func Update(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	vars := mux.Vars(r)
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := CreateOrUpdate(body, vars["projectId"], db); err != nil {
		panic(err)
	}
	return http.StatusOK, JsonResponse{"Updated", true}
}
func SaveOptions(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	return http.StatusOK, JsonResponse{}
}
func Save(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	vars := mux.Vars(r)

	ExportJson(vars["projectId"], "sco", db)
	//ExportJson(vars["projectId"], "glossary", db)

	return http.StatusOK, JsonResponse{"OK", true}
}

/*
Export
*/
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
	if err := ioutil.WriteFile(fmt.Sprintf("C:\\data\\www\\sandbox\\ExtBuilder\\tools\\project_%s_%s.json", projectId, treeType), json, 0644); err != nil {
		panic(err)
	}
}

// TODO: create a way to cleanup un attached rows
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

func SaveBackup(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	return http.StatusOK, JsonResponse{}
}
func SaveSettings(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) (int, interface{}) {
	vars := mux.Vars(r)

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(fmt.Sprintf("C:\\data\\www\\sandbox\\ExtBuilder\\tools\\project_%s_settings.json", vars["projectId"]), body, 0644); err != nil {
		panic(err)
	}
	return http.StatusOK, JsonResponse{"OK", true}
}

/*
Import
*/
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
