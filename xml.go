package main

import (
	"encoding/xml"
	"os"
	"fmt"
)

func main() {

	/*
		type Page struct {
			Title string `xml:"title,attr"`

		}
		type topics struct{
			pages []page
		}
		type Topic struct {

		type Result struct {
			//XMLName xml.Name `xml:"sco"`
			Pages []Page `xml:",any"`
			//pages []page `xml:"topics>page"`
			//topics []topics
		}
		v := Result{}
		data := `
			<sco>
				<topics>
					<page title="[New page Visual Layout page]">asdf</page>
					<page title="[New page Visual Layout page2]">asdf2</page>
				</topics>
			</sco>
		`
		}*/

	type Redirect struct {
		Title string `xml:"title,attr"`
	}

	type Page struct {
		Title string   `xml:"title,attr"`
		//Redir Redirect `xml:"redirect"`
		//Text  string   `xml:"revision>text"`
	}

	xmlFile, _ := os.Open("sco.xml")
	decoder := xml.NewDecoder(xmlFile)

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		//fmt.Printf("Token: %#v\n", t)
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			//fmt.Printf("XMLName: %#v\n", se.Name.Local)
			
			// If we just read a StartElement token
			// ...and its name is "page"
			if se.Name.Local == "page" {
				var p Page
				// decode a whole chunk of following XML into the
				// variable p which is a Page (se above)
				decoder.DecodeElement(&p, &se)
				fmt.Printf("PageTitle: %#v\n", p)
				// Do some stuff with the page.
				//p.Title = CanonicalizeTitle(p.Title)
			}
		}
	}

	/*

		err := xml.Unmarshal([]byte(data), &v)
		if err != nil {
			fmt.Printf("error: %v", err)
			return
		}
		//fmt.Printf("XMLName: %#v\n", v.XMLName)
		fmt.Printf("XMLName: %#v\n", v)
		
		fmt.Printf("XMLName: %#v\n", a)
		//fmt.Printf("pages: %#v\n", v.topics[0].pages)
		//fmt.Printf("pages: %q\n", v.topics)
		//fmt.Printf("Name: %q\n", v.Name)
		//fmt.Printf("Phone: %q\n", v.Phone)
		//fmt.Printf("Email: %v\n", v.Email)
		//fmt.Printf("Groups: %v\n", v.Groups)
		//fmt.Printf("Address: %v\n", v.Address)
	*/
}
