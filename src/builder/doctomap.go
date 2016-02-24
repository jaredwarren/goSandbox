package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

var scoXML = `
<?xml version="1.0"?>
<sco tracking="none" title="as3" narration="true" glossary="true" saturationPercent="50" persIndex="false" courseLogo="" activateTimer="true" timerSeconds="2" quizTimerSeconds="20" activateInactiveTimer="true" timerIntMinutes="5" inactiveResponse="stopRecord" pdfTime="5" reportError="false" hideFluencyTime="false" showHelpOnStart="true" audioButton="true" exitButton="true" pageNumbering="true" include_question_cnt="false" indexTooltip="true" transition="none" tranDirection="right" transitionTime="1" styleName="AS3 Omni Template (Player 10, 1024x768)">
  <topics isthisused="true">
    <page title="[New Page Text and Image]" pType="Text and Image" imageFile="http://unison/access/Media/-Imgur.jpg" imgPos="left" imageWidth="40" nType="none" file="swfs/templateswfs/rapidintake_textimage/ri_textimage_pg.swf" loadPercentage="0" fpsec="12" isTocEntry="true" nonNavPage="false">
	  <pText>&lt;P ALIGN="LEFT"&gt;asdf&lt;/P&gt;</pText>
	</page>
    <topic title="[New Topic]" objectiveId="none" nav="unrestricted" showTopicPages="true">
	    <page title="[New Page Text and Image]" pType="Text and Image" imageFile="http://unison/access/Media/-Imgur.jpg" imgPos="left" imageWidth="40" nType="none" file="swfs/templateswfs/rapidintake_textimage/ri_textimage_pg.swf" loadPercentage="0" fpsec="12" isTocEntry="true" nonNavPage="false">
		  <pText>&lt;P ALIGN="LEFT"&gt;asdf&lt;/P&gt;</pText>
		</page>
    </topic>
  </topics>
</sco>
`

var quizXML = ``
var quizTree *Node
var glossaryXML = ``

// reg
type RegReplace struct {
	find    *regexp.Regexp // XML tag
	replace string         // element value
}

var cleanRegexps = []RegReplace{
	{
		regexp.MustCompile(`(?ms)(<|&lt;)FONT (FACE=("|&quot;)(.+?)("|&quot;))? (SIZE=("|&quot;)(.+?)("|&quot;))? (COLOR=("|&quot;)(.+?)("|&quot;))?`),
		"${1}span style=\"font-family:'$4'; font-size:${8}px; color:$12;\"",
	},

	{
		regexp.MustCompile(`(?ms)(<|&lt;)FONT (COLOR=("|&quot;)(.+?)("|&quot;))?`),
		`${1}span style="color:${4};"`,
	},
	{
		regexp.MustCompile(`TEXTFORMAT`),
		`span`,
	},

	{
		regexp.MustCompile(`/FONT`),
		`/span`,
	},

	{
		regexp.MustCompile(`(?i)align="(.+?)"`),
		`style="text-align:$1"`,
	},

	{
		regexp.MustCompile(`style="text-align:$1"`),
		``,
	},
}

var regsize = regexp.MustCompile(`(?ms)(<|&lt;)FONT (FACE=("|&quot;)(.+?)("|&quot;))? (SIZE=("|&quot;)(.+?)("|&quot;))? (COLOR=("|&quot;)(.+?)("|&quot;))?`)

var regLink = regexp.MustCompile(`(?i)<a (.+?)>`)
var regLinkasf = regexp.MustCompile(`(?i)href="asfunction(.+?)"`)
var regLinktarget = regexp.MustCompile(`(?i)target="_blank"`)

var currentTopicCounter = 1
var currentPageCounter = 1

func main() {
	// load xml
	scoXML = getXml("sco")
	scoTree, merr := DocToTree(scoXML)
	if scoTree == nil || merr != nil {
		panic(merr)
	}

	quizXML = getXml("quiz")
	quizTree, _ = DocToTree(quizXML)

	// settings
	settings := make(map[string]interface{})
	// root
	root := make(map[string]interface{})
	root["id"] = "root"
	root["recordId"] = "root"
	root["parentId"] = nil
	root["leaf"] = false
	root["title"] = "root"
	for _, node := range scoTree.nodes {
		if node.key == "topics" {
			root["page"] = addChildren(node, root, false)
		} else {
			root[node.key] = recast(node.val, true)
			settings[node.key] = recast(node.val, true)
		}
	}
	jsonString, _ := json.Marshal(root)
	jsonString = bytes.Replace(jsonString, []byte("\\u003c"), []byte("<"), -1)
	jsonString = bytes.Replace(jsonString, []byte("\\u003e"), []byte(">"), -1)
	jsonString = bytes.Replace(jsonString, []byte("\\u0026"), []byte("&"), -1)
	// faster hack
	jsonString = bytes.Replace(jsonString, []byte(`"distractorImage"`), []byte(`"distractor"`), -1)

	ioutil.WriteFile(`C:\data\www\sandbox\ExtPlayer\data\sco.json`, jsonString, 0777)

	// write out settings
	jsonString, _ = json.Marshal(settings)
	jsonString = bytes.Replace(jsonString, []byte("\\u003c"), []byte("<"), -1)
	jsonString = bytes.Replace(jsonString, []byte("\\u003e"), []byte(">"), -1)
	jsonString = bytes.Replace(jsonString, []byte("\\u0026"), []byte("&"), -1)
	ioutil.WriteFile(`C:\data\www\sandbox\ExtPlayer\data\settings.json`, jsonString, 0777)
}

func getXml(fileType string) string {
	xmlB, err := ioutil.ReadFile(fmt.Sprintf(`C:\data\www\sandbox\ExtPlayer\%s.xml`, fileType))
	if err != nil {
		panic(err)
	}
	return cleanXmlString(string(xmlB))
}

func cleanXmlString(dirtyXmlString string) string {
	for _, reg := range cleanRegexps {
		dirtyXmlString = reg.find.ReplaceAllString(dirtyXmlString, reg.replace)
	}
	return dirtyXmlString
}

// TODO:
// add quiz!!!
// 	feedback
//  tries
//  distractorImage

var restrictedTopicIds = make([]string, 0)

func addChildren(parent *Node, parentMap map[string]interface{}, showTopicPages bool) []map[string]interface{} {
	children := make([]map[string]interface{}, 0)
	for _, node := range parent.nodes {
		if node.key == "page" {
			pageMap := node.treeToMap(true).(map[string]interface{})
			pageMap["leaf"] = true
			pageMap["parentId"] = parentMap["id"]
			pageMap["completed"] = false
			pageMap["mtype"] = "PageModel"
			pageMap["recordId"] = "PageModel-" + strconv.Itoa(currentPageCounter)
			pageMap["pType"] = cleanPType(pageMap["pType"].(string), pageMap)
			//pageMap["id"] = "PageModel-" + strconv.Itoa(currentPageCounter)
			// append quiz questions
			if quizId, ok := pageMap["id"]; ok {
				var quizNode *Node
				// find quiz
				for _, qn := range quizTree.nodes {
					if qn.Get("id").val == quizId.(string) {
						quizNode = qn
						break
					}
				}
				// add quiz to page
				if quizNode != nil {
					questions := make([]map[string]interface{}, 0)

					for _, qn := range quizNode.nodes {
						if qn.key != "question" {
							if qn.key == "useSupset" {
								pageMap["useSubset"] = qn.val
							} else {
								pageMap[qn.key] = qn.treeToMap(true)
							}
						} else {
							// add question
							question := qn.treeToMap(true).(map[string]interface{})
							// for some reason the id is causing problems...
							delete(question, "id")
							questions = append(questions, question)
						}
					}
					pageMap["question"] = questions
				} else {
					fmt.Println(quizId)
					panic("quiz not found")
				}
			}
			currentPageCounter += 1
			if showTopicPages {
				pageMap["nonNavPage"] = true
			}
			file, ok := pageMap["file"]
			if !ok {
				file = "nofile"
			}
			pageMap["bookmark"] = getBookmark(file.(string))
			if len(restrictedTopicIds) > 0 {
				pageMap[""] = restrictedTopicIds
			}
			children = append(children, pageMap)
		} else if node.key == "topic" {
			pageMap := make(map[string]interface{})
			pageMap["leaf"] = false
			pageMap["topicId"] = currentTopicCounter
			pageMap["parentId"] = parentMap["id"]
			pageMap["mtype"] = "TopicModel"
			pageMap["recordId"] = "TopicModel-" + strconv.Itoa(currentTopicCounter)
			//pageMap["id"] = "TopicModel-" + strconv.Itoa(currentTopicCounter)
			currentTopicCounter += 1

			showtp := showTopicPages
			if node.Get("showTopicPages").val == "false" {
				// If i want to make all topic after this one restricted, change to showTopicPages = true
				showtp = true
			}

			pageMap["page"] = addChildren(node, pageMap, showtp)
			if len(restrictedTopicIds) > 0 {
				pageMap[""] = restrictedTopicIds
			}

			if node.Get("nav").val == "complete" {
				restrictedTopicIds = append(restrictedTopicIds, pageMap["topicId"].(string))
			}

			children = append(children, pageMap)
		} else {
			parentMap[node.key] = recast(node.val, true)
		}
	}
	return children
}

var pTypeReg = regexp.MustCompile(` `)

func cleanPType(pType string, node map[string]interface{}) string {
	pType = pTypeReg.ReplaceAllString(pType, "")
	if pType == "ExternalSWFdefault" || pType == "CustomContent" || pType == "CustomHTML5Page" || pType == "default" || pType == "PowerPointSlide" {
		if node["sourceType"] == "upload" {
			pType = "FrameHTML5Page"
		} else {
			pType = "CustomHTML5Page"
		}
	}
	if pType == "WordGameNon-Timed" {
		pType = "WordGame"
		node["timed"] = false
	}
	if pType == "WordGameTimed" {
		pType = "WordGame"
		node["timed"] = true
	}
	if pType == "quiz" {
		pType = "Quiz"
	}

	if pType == "StreamingVideo" {
		pType = "Video"
	}

	switch pType {
	case "TextImageLinkandAudio":
		return "TextPage"
	case "TextImageandAudio":
		return "TextPage"
	case "TextandAudio":
		return "TextPage"
	case "TextandImageLink":
		return "TextPage"
	case "TextandImage":
		return "TextPage"
	case "Text":
		return "TextPage"
	case "ImageLinkandAudio":
		return "TextPage"
	case "ImageandAudio":
		return "TextPage"
	default:
		break
	}

	return pType
}

// bookmark
var bookmark = make(map[string]int)

func getBookmark(file string) string {
	count := 0
	if currentCount, ok := bookmark[file]; ok {
		count = currentCount
		bookmark[file] = count + 1
	} else {
		bookmark[file] = count + 1
	}
	return file + "|" + strconv.Itoa(count)
}

//-----------------------------
// https://github.com/clbanning/x2j/blob/master/x2j.go below

type Node struct {
	dup   bool   // is member of a list
	attr  bool   // is an attribute
	key   string // XML tag
	val   string // element value
	nodes []*Node
}

// DocToJson - return an XML doc as a JSON string.
//	If the optional argument 'recast' is 'true', then values will be converted to boolean or float64 if possible.
func DocToJson(doc string, recast ...bool) (string, error) {
	var r bool
	if len(recast) == 1 {
		r = recast[0]
	}
	m, merr := DocToMap(doc, r)
	if m == nil || merr != nil {
		return "", merr
	}

	b, berr := json.Marshal(m)
	if berr != nil {
		return "", berr
	}

	// NOTE: don't have to worry about safe JSON marshaling with json.Marshal, since '<' and '>" are reservedin XML.
	return string(b), nil
}

// DocToJsonIndent - return an XML doc as a prettified JSON string.
//	If the optional argument 'recast' is 'true', then values will be converted to boolean or float64 if possible.
//	Note: recasting is only applied to element values, not attribute values.
func DocToJsonIndent(doc string, recast ...bool) (string, error) {
	var r bool
	if len(recast) == 1 {
		r = recast[0]
	}
	m, merr := DocToMap(doc, r)
	if m == nil || merr != nil {
		return "", merr
	}

	b, berr := json.MarshalIndent(m, "", "  ")
	if berr != nil {
		return "", berr
	}

	// NOTE: don't have to worry about safe JSON marshaling with json.Marshal, since '<' and '>" are reservedin XML.
	return string(b), nil
}

// DocToMap - convert an XML doc into a map[string]interface{}.
// (This is analogous to unmarshalling a JSON string to map[string]interface{} using json.Unmarshal().)
//	If the optional argument 'recast' is 'true', then values will be converted to boolean or float64 if possible.
//	Note: recasting is only applied to element values, not attribute values.
func DocToMap(doc string, recast ...bool) (map[string]interface{}, error) {
	var r bool
	if len(recast) == 1 {
		r = recast[0]
	}
	n, err := DocToTree(doc)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	m[n.key] = n.treeToMap(r)

	return m, nil
}

// DocToTree - convert an XML doc into a tree of nodes.
func DocToTree(doc string) (*Node, error) {
	// xml.Decoder doesn't properly handle whitespace in some doc
	// see songTextString.xml test case ...
	reg, _ := regexp.Compile("[ \t\n\r]*<")
	doc = reg.ReplaceAllString(doc, "<")

	b := bytes.NewBufferString(doc)
	p := xml.NewDecoder(b)
	n, berr := xmlToTree("", nil, p)
	if berr != nil {
		return nil, berr
	}

	return n, nil
}

// (*Node)WriteTree - convert a tree of nodes into a printable string.
//	'padding' is the starting indentation; typically: n.WriteTree().
func (n *Node) WriteTree(padding ...int) string {
	var indent int
	if len(padding) == 1 {
		indent = padding[0]
	}

	var s string
	if n.val != "" {
		for i := 0; i < indent; i++ {
			s += "  "
		}
		s += n.key + " : " + n.val + "\n"
	} else {
		for i := 0; i < indent; i++ {
			s += "  "
		}
		s += n.key + " :" + "\n"
		for _, nn := range n.nodes {
			s += nn.WriteTree(indent + 1)
		}
	}
	return s
}

// get node by key
func (n *Node) Get(key string) *Node {
	for _, node := range n.nodes {
		if node.key == key {
			return node
		}
	}
	return nil
}
func (n *Node) GetByVal(key string, val string) *Node {
	for _, node := range n.nodes {
		if node.key == key && node.val == val {
			return node
		}
	}
	return nil
}

// xmlToTree - load a 'clean' XML doc into a tree of *Node.
func xmlToTree(skey string, a []xml.Attr, p *xml.Decoder) (*Node, error) {
	n := new(Node)
	n.nodes = make([]*Node, 0)

	if skey != "" {
		n.key = skey
		if len(a) > 0 {
			for _, v := range a {
				na := new(Node)
				na.attr = true
				na.key = v.Name.Local
				na.val = v.Value
				n.nodes = append(n.nodes, na)
			}
		}
	}
	for {
		t, err := p.Token()
		if err != nil {
			return nil, err
		}
		switch t.(type) {
		case xml.StartElement:
			tt := t.(xml.StartElement)
			// handle root
			if n.key == "" {
				n.key = tt.Name.Local
				if len(tt.Attr) > 0 {
					for _, v := range tt.Attr {
						na := new(Node)
						na.attr = true
						na.key = v.Name.Local
						na.val = v.Value
						n.nodes = append(n.nodes, na)
					}
				}
			} else {
				nn, nnerr := xmlToTree(tt.Name.Local, tt.Attr, p)
				if nnerr != nil {
					return nil, nnerr
				}
				n.nodes = append(n.nodes, nn)
			}
		case xml.EndElement:
			// scan n.nodes for duplicate n.key values
			n.markDuplicateKeys()
			return n, nil
		case xml.CharData:
			tt := string(t.(xml.CharData))
			// fix links
			tt = regLink.ReplaceAllStringFunc(tt, func(matches string) string {
				if regLinkasf.MatchString(matches) {
					return regLinktarget.ReplaceAllString(matches, "")
				}
				return matches
			})
			// force all text
			nn := new(Node)
			nn.key = "#text"
			nn.val = tt
			n.nodes = append(n.nodes, nn)
			/*if len(n.nodes) > 0 {
				nn := new(Node)
				nn.key = "#text"
				nn.val = tt
				n.nodes = append(n.nodes, nn)
			} else {
				n.val = tt
			}*/
		default:
			// noop
		}
	}
	// Logically we can't get here, but provide an error message anyway.
	return nil, errors.New("Unknown parse error in xmlToTree() for: " + n.key)
}

// (*Node)markDuplicateKeys - set node.dup flag for loading map[string]interface{}.
func (n *Node) markDuplicateKeys() {
	l := len(n.nodes)
	for i := 0; i < l; i++ {
		if n.nodes[i].dup {
			continue
		}
		for j := i + 1; j < l; j++ {
			if n.nodes[i].key == n.nodes[j].key {
				n.nodes[i].dup = true
				n.nodes[j].dup = true
			}
		}
	}
}

// (*Node)treeToMap - convert a tree of nodes into a map[string]interface{}.
//	(Parses to map that is structurally the same as from json.Unmarshal().)
// Note: root is not instantiated; call with: "m[n.key] = n.treeToMap(recast)".
func (n *Node) treeToMap(r bool) interface{} {
	if len(n.nodes) == 0 {
		return recast(n.val, r)
	}

	m := make(map[string]interface{}, 0)
	for _, v := range n.nodes {
		// just a value
		if !v.dup && len(v.nodes) == 0 {
			m[v.key] = recast(v.val, r)
			continue
		}

		// a list of values
		if v.dup {
			var a []interface{}
			if vv, ok := m[v.key]; ok {
				a = vv.([]interface{})
			} else {
				a = make([]interface{}, 0)
			}
			a = append(a, v.treeToMap(r))
			m[v.key] = interface{}(a)
			continue
		}

		// it's a unique key
		m[v.key] = v.treeToMap(r)
	}

	return interface{}(m)
}

// recast - try to cast string values to bool or float64
var numReg = regexp.MustCompile(`^[0-9]+$`)
var trueReg = regexp.MustCompile(`true`)
var falseReg = regexp.MustCompile(`false`)
var colorReg = regexp.MustCompile(`^0x([0-9A-Fa-f]+)$`)

func recast(s string, r bool) interface{} {
	if r {
		// handle numeric strings ahead of boolean
		if f, err := strconv.ParseUint(s, 10, 64); err == nil {
			return interface{}(f)
		}
		// handle numeric strings ahead of boolean
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return interface{}(f)
		}
		// ParseBool treats "1"==true & "0"==false
		if b, err := strconv.ParseBool(s); err == nil {
			return interface{}(b)
		}
		if trueReg.MatchString(s) {
			return interface{}(true)
		}
		if falseReg.MatchString(s) {
			return interface{}(false)
		}
		if colorReg.MatchString(s) {
			return interface{}("#" + colorReg.ReplaceAllString(s, "$1"))
		}
	}
	return interface{}(s)
}

// WriteMap - dumps the map[string]interface{} for examination.
//	'offset' is initial indentation count; typically: WriteMap(m).
//	NOTE: with XML all element types are 'string'.
//	But code written as generic for use with maps[string]interface{} values from json.Unmarshal().
//	Or it can handle a DocToMap(doc,true) result where values have be recast'd.
func WriteMap(m interface{}, offset ...int) string {
	var indent int
	if len(offset) == 1 {
		indent = offset[0]
	}

	var s string
	switch m.(type) {
	case nil:
		return "[nil] nil"
	case string:
		return "[string] " + m.(string)
	case float64:
		return "[float64] " + strconv.FormatFloat(m.(float64), 'e', 2, 64)
	case bool:
		return "[bool] " + strconv.FormatBool(m.(bool))
	case []interface{}:
		s += "[[]interface{}]"
		for i, v := range m.([]interface{}) {
			s += "\n"
			for i := 0; i < indent; i++ {
				s += "  "
			}
			s += "[item: " + strconv.FormatInt(int64(i), 10) + "]"
			switch v.(type) {
			case string, float64, bool:
				s += "\n"
			default:
				// noop
			}
			for i := 0; i < indent; i++ {
				s += "  "
			}
			s += WriteMap(v, indent+1)
		}
	case map[string]interface{}:
		for k, v := range m.(map[string]interface{}) {
			s += "\n"
			for i := 0; i < indent; i++ {
				s += "  "
			}
			// s += "[map[string]interface{}] "+k+" :"+WriteMap(v,indent+1)
			s += k + " :" + WriteMap(v, indent+1)
		}
	default:
		// shouldn't ever be here ...
		s += fmt.Sprintf("unknown type for: %v", m)
	}
	return s
}

// ------------------------  value extraction from XML doc --------------------------

// DocValue - return a value for a specific tag
//	'doc' is a valid XML message.
//	'path' is a hierarchy of XML tags, e.g., "doc.name".
//	'attrs' is an optional list of "name:value" pairs for attributes.
//	Note: 'recast' is not enabled here. Use DocToMap(), NewAttributeMap(), and MapValue() calls for that.
func DocValue(doc, path string, attrs ...string) (interface{}, error) {
	n, err := DocToTree(doc)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	m[n.key] = n.treeToMap(false)

	a, aerr := NewAttributeMap(attrs...)
	if aerr != nil {
		return nil, aerr
	}
	v, verr := MapValue(m, path, a)
	if verr != nil {
		return nil, verr
	}
	return v, nil
}

// MapValue - retrieves value based on walking the map, 'm'.
//	'm' is the map value of interest.
//	'path' is a period-separated hierarchy of keys in the map.
//	'attr' is a map of attribute "name:value" pairs from NewAttributeMap().
//	If the path can't be traversed, an error is returned.
//	Note: the optional argument 'r' can be used to coerce attribute values, 'attr', if done so for 'm'.
func MapValue(m map[string]interface{}, path string, attr map[string]interface{}, r ...bool) (interface{}, error) {
	// attribute values may have been recasted during map construction; default is 'false'.
	if len(r) == 1 && r[0] == true {
		for k, v := range attr {
			attr[k] = recast(v.(string), true)
		}
	}

	// parse the path
	keys := strings.Split(path, ".")

	// initialize return value to 'm' so a path of "" will work correctly
	var v interface{} = m
	var ok bool
	var okey string
	var isMap bool = true
	if keys[0] == "" {
		goto checkAttr
	}
	for _, key := range keys {
		if !isMap {
			return nil, errors.New("no keys beyond: " + okey)
		}
		if v, ok = m[key]; !ok {
			return nil, errors.New("no key in map: " + key)
		} else {
			switch v.(type) {
			case map[string]interface{}:
				m = v.(map[string]interface{})
				isMap = true
			default:
				isMap = false
			}
		}
		// save 'key' for error reporting
		okey = key
	}

checkAttr:
	// no attributes to match, then done
	if len(attr) == 0 {
		return v, nil
	}

	// match attributes; value is "#text" or nil
	return hasAttributes(v, attr)
}

// hasAttributes() - interface{} equality works for string, float64, bool
func hasAttributes(v interface{}, a map[string]interface{}) (interface{}, error) {
	switch v.(type) {
	case []interface{}:
		// run through all entries looking one with matching attributes
		for _, vv := range v.([]interface{}) {
			if vvv, vvverr := hasAttributes(vv, a); vvverr == nil {
				return vvv, nil
			}
		}
		return nil, errors.New("no list member with matching attributes")
	case map[string]interface{}:
		// do all attribute name:value pairs match?
		nv := v.(map[string]interface{})
		for key, val := range a {
			if vv, ok := nv[key]; !ok {
				return nil, errors.New("no attribute with name: " + key[1:])
			} else if val != vv {
				return nil, errors.New("no attribute key:value pair: " + fmt.Sprintf("%s:%v", key[1:], val))
			}
		}
		// they all match; so return value associated with "#text" key.
		if vv, ok := nv["#text"]; ok {
			return vv, nil
		} else {
			// this happens when another element is value of tag rather than just a string value
			return nv, nil
		}
	}
	return nil, errors.New("no match for attributes")
}

// NewAttributeMap() - generate map of attributes=value entries as map["-"+string]string.
//	'kv' arguments are "name:value" pairs that appear as attributes, name="value".
func NewAttributeMap(kv ...string) (map[string]interface{}, error) {
	m := make(map[string]interface{}, 0)
	for _, v := range kv {
		vv := strings.Split(v, ":")
		if len(vv) != 2 {
			return nil, errors.New("attribute not \"name:value\" pair: " + v)
		}
		// attributes are stored as keys prepended with hyphen
		m["-"+vv[0]] = interface{}(vv[1])
	}
	return m, nil
}

//------------------------- get values for key ----------------------------

// ValuesForTag - return all values in doc associated with 'tag'.
//	Returns nil if the 'tag' does not occur in the doc.
//	If there is an error encounted while parsing doc, that is returned.
//	If you want values 'recast' use DocToMap() and ValuesForKey().
func ValuesForTag(doc, tag string) ([]interface{}, error) {
	n, err := DocToTree(doc)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	m[n.key] = n.treeToMap(false)

	return ValuesForKey(m, tag), nil
}

// ValuesForKey - return all values in map associated with 'key'
//	Returns nil if the 'key' does not occur in the map
func ValuesForKey(m map[string]interface{}, key string) []interface{} {
	ret := make([]interface{}, 0)

	hasKey(m, key, &ret)
	if len(ret) > 0 {
		return ret
	}
	return nil
}

// hasKey - if the map 'key' exists append it to array
//          if it doesn't do nothing except scan array and map values
func hasKey(iv interface{}, key string, ret *[]interface{}) {
	switch iv.(type) {
	case map[string]interface{}:
		vv := iv.(map[string]interface{})
		if v, ok := vv[key]; ok {
			*ret = append(*ret, v)
		}
		for _, v := range iv.(map[string]interface{}) {
			hasKey(v, key, ret)
		}
	case []interface{}:
		for _, v := range iv.([]interface{}) {
			hasKey(v, key, ret)
		}
	}
}
