package speed

import (
	"bytes"
	"fmt"
	"log"

	"golang.org/x/net/html"
)

type Tag struct {
	Name string
	Text [][]string
	Attr [][]Attribute
	Node [][]Node
}

type TagList struct {
	Tags map[string][]Tag
}

// ultra specific search struct
type SearchSpecifc struct {
	Tag  string
	Attr map[string]string
}

type SearchGroup struct {
	Group []Search
}

type Doc struct {
	Doc      *html.Node
	Tags     TagList
	NodeName string
}

// type Selectors struct {
// 	Type string
// 	Name string
// }

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

func GetTags(r *html.Node, s SearchSpecifc, l *Tag) {

	getTags(r, s, l)
}

func getTags(r *html.Node, s SearchSpecifc, l *Tag) *html.Node {

	isEmpty := false
	if !(len(s.Tag) > 0) {
		isEmpty = true
	}

	var f func(x *html.Node)
	var f_TagEmpty func(r *html.Node)

	f = func(x *html.Node) {
		if x.Type == html.ElementNode && x.Data == s.Tag {
			temp := []Attribute{}
			for i := 0; i < len(x.Attr); i++ {
				attr := x.Attr[i]
				if compareAttrandValue(attr, s) || s.Attr == nil {
					temp = append(temp, Attribute{Name: attr.Key, Value: attr.Val})
					addInnerHtml(x, l)
				}
			}

			if len(temp) > 0 {
				l.Attr = append(l.Attr, temp)
				temp = nil
			}

		}

		for c := x.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f_TagEmpty = func(r *html.Node) {
		if r.Type == html.ElementNode {
			temp := []Attribute{}
			for i := 0; i < len(r.Attr); i++ {
				attr := r.Attr[i]
				if compareAttrandValue(attr, s) {
					temp = append(temp, Attribute{Name: attr.Key, Value: attr.Val})
					addInnerHtml(r, l)
				}
			}
			if len(temp) > 0 {
				l.Attr = append(l.Attr, temp)
				temp = nil
			}
		}
		for c := r.FirstChild; c != nil; c = c.NextSibling {
			f_TagEmpty(c)
		}
	}

	if isEmpty {
		f_TagEmpty(r)
	} else {
		f(r)
	}

	return r
}

func addInnerHtml(r *html.Node, l *Tag) {
	b := &bytes.Buffer{}
	collectText(r, b)
	l.Text = append(l.Text, []string{b.String()})
}

func collectText(r *html.Node, b *bytes.Buffer) {
	if r.Type == html.TextNode {
		b.WriteString(r.Data)
	}
	for c := r.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, b)
	}
}

// func renderNode(r *html.Node) string {
// 	var buf bytes.Buffer
// 	w := io.Writer(&buf)
// 	html.Render(w, r)
// 	return buf.String()
// }
