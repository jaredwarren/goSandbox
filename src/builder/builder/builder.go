package builder

import (
	"builder/ini"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	/*if err := json.NewEncoder(w).Encode(nodes); err != nil {
		panic(err)
	}*/
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
func Create(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}
func Destroy(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}
func Update(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}
func Save(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}
func SaveBackup(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}
func SaveSettings(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}

// Import
func Import(w http.ResponseWriter, r *http.Request, db *sql.DB, config *ini.Dict) {

}

func ImportJson(filePath string, db *sql.DB) {
	file, e := ioutil.ReadFile(filePath)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	/*var jsonMap Node
	json.Unmarshal(file, &jsonMap)*/

	//jsonToRow(jsonMap)
	// Creating the maps for JSON
	m := map[string]interface{}{}

	// Parsing/Unmarshalling JSON encoding/json
	err := json.Unmarshal(file, &m)

	if err != nil {
		panic(err)
	}
	parseMap(m)
}

func jsonToRow(jsonMap Node) []NodeRow {
	rows := make([]NodeRow, 0)
	for key, value := range jsonMap.Children {
		fmt.Printf("%v:%T\n", key, value)
	}

	return rows
}
func parseMap(aMap map[string]interface{}) {
	for key, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			// pText
			//fmt.Println(key + "---")
			parseMap(val.(map[string]interface{}))
		case []interface{}:
			// page
			//fmt.Println(key + ":::")
			parseArray(val.([]interface{}))
		default:
			fmt.Println(key)
			if concreteVal == "" {

			}
			// other
			//fmt.Println(key, ":", concreteVal)
		}
	}
}

func parseArray(anArray []interface{}) {
	for i, val := range anArray {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			// List of pages
			parseMap(val.(map[string]interface{}))
		case []interface{}:
			// not used???
			fmt.Println("Index:", i)
			parseArray(val.([]interface{}))
		default:
			// not used???
			fmt.Println("Index:::", i, ":", concreteVal)

		}
	}
}
