package speed

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/html"
	// "strings"
)

var (
	// dfs func(x *html.Node, data string)

	MinHeaderList = make(map[string]string)
	HeaderList    = make([]Headers, 0)
	DocumentList  = make([]Doc, 0)
	// TagTrackerList = make([]TagTracker, 0)
)

type Headers struct {
	UserAgent               string
	Accept                  string
	AcceptLanguage          string
	AcceptEncoding          string
	Referer                 string
	Connection              string
	UpgradeInsecureRequests string
	IfModifiedSince         string
	IfNoneMatch             string
	CacheControl            string
}

type Doc struct {
	Doc      *html.Node
	Tags     TagList
	NodeName string
}

type Attribute struct {
	name  string
	value string
}

type Tag struct {
	Text       string
	Attributes []Attribute
}

type TagList struct {
	Tags map[string][]Tag
}

// works with the bytestream that naturally comes out of the Parse function
// adds it to the document list
func SetupDocument(bytestream *[]byte) Doc {
	b := *bytestream
	r, err := Parse(bytes.NewReader(b))
	if err != nil {
		log.Fatal("Failed to PARSE")
	}

	return AddtoDocument(r)
}

func AddtoDocument(r *html.Node) Doc {
	doc := Doc{Doc: r, NodeName: r.Data}
	DocumentList = append(DocumentList, doc)

	return doc
}

// returns bytes. needs to be converted to string in order to be readable to users
func Get(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Failed to get LINK")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to READ")
	}
	return body
}

// Returns a list of tags and its data when found
// does depth first search
func SearchTag(h *html.Node, tag string) {
	fmt.Println(h.Data)
	for c := h.FirstChild; c != nil; c = c.NextSibling {
		SearchTag(c, tag)
	}
}

func (r *Doc) SearchTag(tag string) []Tag {

	var dfs func(x *html.Node, data string)

	t := make([]Tag, 0)

	dfs = func(x *html.Node, data string) {
		// if x.Data == tag {
		// 	t = append(t, Tag{Text: x.Data})
		// }
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			dfs(c, data)
		}
	}

	dfs(r.Doc, tag)

	return t
}

// func (r *Doc) AddTag(t Tag) {
// 	r.Tags = append(r.Tags, t)
// }

// returns utf-8 string text
func GetHtml(url string) string {
	return string(Get(url))
}

func Parse(r io.Reader) (*html.Node, error) {
	return html.Parse(r)
}

func EasyParse(s *[]byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(*s))
}
