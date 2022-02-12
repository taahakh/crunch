package speed

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"golang.org/x/net/html"
)

type Doc struct {
	Doc      *html.Node
	Tags     TagList
	NodeName string
}

type Attribute struct {
	Name  string
	Value string
}

type Tag struct {
	Text string
	Attr [][]Attribute
}

type TagList struct {
	Tags map[string][]Tag
}

type Search struct {
	Tag  string
	Attr map[string]string
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

// Returns a list of tags and its data when found
// does depth first search
func SearchTag(h *html.Node, tag string) {
	if h.Type == html.ElementNode && h.Data == tag {
		fmt.Println("This tag has been found!")
	}
	// fmt.Println(h.Data)
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

func FindAll(r *html.Node, t string, l *Tag) (*html.Node, bool) {
	if r.Type == html.ElementNode && r.Data == t {
		tmp := []Attribute{}
		for i := 0; i < len(r.Attr); i++ {
			attr := r.Attr[i]
			tmp = append(tmp, Attribute{Name: attr.Key, Value: attr.Val})
		}

		l.Attr = append(l.Attr, tmp)

		tmp = nil
	}
	for c := r.FirstChild; c != nil; c = c.NextSibling {
		FindAll(c, t, l)
	}

	return r, true
}

func GetTags(r *html.Node, s Search, l *Tag) *html.Node {
	if r.Type == html.ElementNode && r.Data == s.Tag {
		for i := 0; i < len(r.Attr); i++ {
			attr := r.Attr[i]
			if compareAttrandValue(attr, s) {
				l.Attr = append(l.Attr, []Attribute{{Name: attr.Key, Value: attr.Val}})
			}
		}
	}

	for c := r.FirstChild; c != nil; c = c.NextSibling {
		GetTags(c, s, l)
	}

	return r
}

func compareAttrandValue(attr html.Attribute, s Search) bool {
	if val, ok := s.Attr[attr.Key]; ok {
		fmt.Print(" ", attr.Key)
		fmt.Println(" ", attr.Val)
		if attr.Val == val {
			return true
		}
	}
	return false
}

func Parse(r io.Reader) (*html.Node, error) {
	return html.Parse(r)
}

func EasyParse(s *[]byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(*s))
}
