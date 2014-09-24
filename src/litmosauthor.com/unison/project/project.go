package project

import (
	"database/sql"

	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
	//"reflect"
)

type Tags struct {
	Id   int
	Name string
}

type Content struct {
	Id      int
	Title   string
	Content string
}

type Comment struct {
	Id   int
	Note string
}

type Page struct {
	Tags     *Tags
	Content  *Content
	Comment  *Comment
	Projects Projects
}

type Projects []Project

func (p Projects) HasProjects() bool {
	return len(p) > 0
}

type Project struct {
	Id     string `db:"project_id"`
	Name   string `db:"project_name"`
	CustId string `db:"cust_id"`
}

func NewProject() *Project {
	return &Project{}
}

func getProjects(db *sql.DB) Projects {
	projects := Projects{}

	cust_id := "unison"
	rows, err := db.Query("SELECT project_id, project_name, cust_id FROM project WHERE cust_id=?", cust_id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		//var project Project
		project := Project{}
		if err := rows.Scan(&project.Id, &project.Name, &project.CustId); err != nil {
			log.Fatal(err)
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return projects
}

func Dashboard(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("Project::Dashboard")

	projects := getProjects(db)
	//fmt.Println(projects.HasProjects())

	// for now parse every request so I don't have to recompile, maybe
	tmpl := make(map[string]*template.Template)
	tmpl["index.html"] = template.Must(template.ParseFiles("static/templates/index.html", "static/templates/content.html", "static/templates/base.html"))
	tmpl["other.html"] = template.Must(template.ParseFiles("static/templates/other.html", "static/templates/base.html"))

	pagedata := &Page{Tags: &Tags{Id: 1, Name: "golang"},
		Content:  &Content{Id: 9, Title: "Hello", Content: "World!"},
		Projects: projects,
		Comment:  &Comment{Id: 2, Note: "Good Day!"}}

	tmpl["index.html"].ExecuteTemplate(w, "base", pagedata)
}

var templates = template.Must(template.ParseFiles("static/templates/404.html", "static/templates/home.html"))

func executeTemplate(w http.ResponseWriter, tmpl string) {
	fmt.Printf("Project Template:" + tmpl)
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// scanAll scans all rows into a destination, which must be a slice of any
// type.  If the destination slice type is a Struct, then StructScan will be
// used on each row.  If the destination is some other kind of base type, then
// each row must only have one column which can scan into that type.  This
// allows you to do something like:
//
//    rows, _ := db.Query("select id from people;")
//    var ids []int
//    scanAll(rows, &ids, false)
//
// and ids will be a list of the id results.  I realize that this is a desirable
// interface to expose to users, but for now it will only be exposed via changes
// to `Get` and `Select`.  The reason that this has been implemented like this is
// this is the only way to not duplicate reflect work in the new API while
// maintaining backwards compatibility.
/*func scanAll(rows *Rows, dest interface{}, structOnly bool) error {
	var v, vp reflect.Value

	value := reflect.ValueOf(dest)

	// json.Unmarshal returns errors for these
	if value.Kind() != reflect.Ptr {
		//return errors.New("must pass a pointer, not a value, to StructScan destination")
		log.Fatal("must pass a pointer, not a value, to StructScan destination")
	}
	if value.IsNil() {
		log.Fatal("nil pointer passed to StructScan destination")
		//return errors.New("nil pointer passed to StructScan destination")
	}
	direct := reflect.Indirect(value)

	slice, err := baseType(value.Type(), reflect.Slice)
	if err != nil {
		return err
	}

	isPtr := slice.Elem().Kind() == reflect.Ptr
	base := reflectx.Deref(slice.Elem())
	isStruct := base.Kind() == reflect.Struct
	// check if a pointer to the slice type implements sql.Scanner; if it does, we
	// will treat this as a base type slice rather than a struct slice;  eg, we will
	// treat []sql.NullString as a single row rather than a struct with 2 scan targets.
	isScanner := reflect.PtrTo(base).Implements(_scannerInterface)

	// if we must have a struct and the base type isn't a struct, return an error.
	// this maintains API compatibility for StructScan, which is only important
	// because StructScan should involve structs and it feels gross to add more
	// weird junk to it.
	if structOnly {
		if !isStruct {
			return fmt.Errorf("expected %s but got %s", reflect.Struct, base.Kind())
		}
		if isScanner {
			return fmt.Errorf("structscan expects a struct dest but the provided struct type %s implements scanner", base.Name())
		}
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// if it's a base type make sure it only has 1 column;  if not return an error
	if !isStruct && len(columns) > 1 {
		return fmt.Errorf("non-struct dest type %s with >1 columns (%d)", base.Kind(), len(columns))
	}

	if isStruct && !isScanner {
		var values []interface{}
		var m *reflectx.Mapper

		switch rows.(type) {
		case *Rows:
			m = rows.(*Rows).Mapper
		default:
			m = mapper()
		}

		fields := m.TraversalsByName(base, columns)
		// if we are not unsafe and are missing fields, return an error
		if f, err := missingFields(fields); err != nil && !isUnsafe(rows) {
			return fmt.Errorf("missing destination name %s", columns[f])
		}
		values = make([]interface{}, len(columns))

		for rows.Next() {
			// create a new struct type (which returns PtrTo) and indirect it
			vp = reflect.New(base)
			v = reflect.Indirect(vp)

			err = fieldsByTraversal(v, fields, values, true)

			// scan into the struct field pointers and append to our results
			err = rows.Scan(values...)
			if err != nil {
				return err
			}

			if isPtr {
				direct.Set(reflect.Append(direct, vp))
			} else {
				direct.Set(reflect.Append(direct, v))
			}
		}
	} else {
		for rows.Next() {
			vp = reflect.New(base)
			err = rows.Scan(vp.Interface())
			// append
			if isPtr {
				direct.Set(reflect.Append(direct, vp))
			} else {
				direct.Set(reflect.Append(direct, reflect.Indirect(vp)))
			}
		}
	}

	return rows.Err()
}

func baseType(t reflect.Type, expected reflect.Kind) (reflect.Type, error) {
	t = reflectx.Deref(t)
	if t.Kind() != expected {
		return nil, fmt.Errorf("expected %s but got %s", expected, t.Kind())
	}
	return t, nil
}
*/
