package speed

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

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
	Name string
	Text [][]string
	Attr [][]Attribute
}

type TagList struct {
	Tags map[string][]Tag
}

type SearchSpecifc struct {
	Tag  string
	Attr map[string]string
}

type Search struct {
	Tag      string
	Selector []Selectors
	Attr     []string
}

type SearchGroup struct {
	Group []Search
}

type SearchBuilder struct {
	ANstart, Bracket, FinishBracket bool
	LeftStart, SelectorState        bool
	Tracking                        []string
	Attr                            []string
	Selector                        []Selectors
	Tag                             string
}

type Selectors struct {
	Type string
	Name string
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// func CreateSearch(s *SearchBuilder) Search {

// }

func FinderParser(s string) {
	lol := []string{}
	for _, v := range s {
		lol = append(lol, string(v))
	}
	build := &SearchBuilder{}
	build.ANstart = false
	build.Bracket = false
	build.FinishBracket = false
	build.LeftStart = false
	build.SelectorState = false
	finderParser(lol, 0, build)
	if len(build.Tracking) > 0 {
		build.Tag = strings.Join(build.Tracking, "")
		build.Tracking = nil
	}
	fmt.Println(build)

}

func finderParser(r []string, i int, b *SearchBuilder) {
	i++
	if i != len(r) {
		finderParser(r, i, b)
	}
	checkStringParse(r[i-1], b, i-1)
}

func checkStringParse(r string, b *SearchBuilder, i int) {

	var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

	var appendstring = func(s []string) string {
		join := strings.Join(s, "")
		return join
	}

	if b.FinishBracket || (!b.Bracket && !b.FinishBracket) {

		if b.FinishBracket {
			if !b.LeftStart {
				if isAlphaNumeric(r) {
					b.Tracking = append(b.Tracking, r)
					b.LeftStart = true
					return
				}
				log.Fatal("not right: left")
			}
		}

		if r == "." || r == "#" {
			if b.SelectorState {
				log.Fatal("selector state")
				return
			}
			b.Selector = append(b.Selector, Selectors{Type: r, Name: Reverse(appendstring(b.Tracking))})
			b.Tracking = nil
			b.SelectorState = true
			return
		}

		if isAlphaNumeric(r) {
			b.SelectorState = false
			b.Tracking = append(b.Tracking, r)
			return
		}

	}

	if r == "]" {
		if b.ANstart || b.Bracket {
			log.Fatal("right bracket")
		}
		if !b.Bracket {
			b.Bracket = true
		}
		return
	}

	if r == "[" {
		// fmt.Print("Im here")
		if !b.Bracket || b.FinishBracket {
			log.Fatal("left bracket")
		}
		b.FinishBracket = true
		b.Attr = append(b.Attr, Reverse(appendstring(b.Tracking)))
		b.Tracking = nil
		return
	}

	if r == "," {
		if !b.Bracket {
			log.Fatal("comma")
		}
		b.Attr = append(b.Attr, Reverse(appendstring(b.Tracking)))
		b.Tracking = nil
		return
	}

	if isAlphaNumeric(r) {
		if !b.Bracket {
			b.ANstart = true
		}
		b.Tracking = append(b.Tracking, r)
		return
	} else {
		log.Fatal("final alphanumeric")
	}
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

func GetTags(r *html.Node, s SearchSpecifc, l *Tag) {

	getTags(r, s, l)
}

// func getTags(r *html.Node, s Search, l *Tag) *html.Node {

// 	isEmpty := false
// 	if !(len(s.Tag) > 0) {
// 		isEmpty = true
// 	}

// 	if r.Type == html.ElementNode && (r.Data == s.Tag || isEmpty) {
// 		temp := []Attribute{}
// 		for i := 0; i < len(r.Attr); i++ {
// 			attr := r.Attr[i]
// 			if compareAttrandValue(attr, s) || len(s.Tag) > 0 {
// 				temp = append(temp, Attribute{Name: attr.Key, Value: attr.Val})
// 				addInnerHtml(r, l)
// 			}
// 		}

// 		if len(temp) > 0 {
// 			l.Attr = append(l.Attr, temp)
// 			temp = nil
// 		}

// 	}

// 	for c := r.FirstChild; c != nil; c = c.NextSibling {
// 		getTags(c, s, l)
// 	}

// 	return r
// }

func getTags(r *html.Node, s SearchSpecifc, l *Tag) *html.Node {

	isEmpty := false
	if !(len(s.Tag) > 0) {
		isEmpty = true
	}

	var f func(x *html.Node)
	var f_TagEmpty func(r *html.Node)

	f = func(x *html.Node) {
		// fmt.Println("got atg")
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

	// if r.Type == html.ElementNode && (r.Data == s.Tag || isEmpty) {
	// 	temp := []Attribute{}
	// 	for i := 0; i < len(r.Attr); i++ {
	// 		attr := r.Attr[i]
	// 		if compareAttrandValue(attr, s) || len(s.Tag) > 0 {
	// 			temp = append(temp, Attribute{Name: attr.Key, Value: attr.Val})
	// 			addInnerHtml(r, l)
	// 		}
	// 	}

	// 	if len(temp) > 0 {
	// 		l.Attr = append(l.Attr, temp)
	// 		temp = nil
	// 	}

	// }

	// for c := r.FirstChild; c != nil; c = c.NextSibling {
	// 	getTags(c, s, l)
	// }

	return r
}

func addInnerHtml(r *html.Node, l *Tag) {
	b := &bytes.Buffer{}
	collectText(r, b)
	l.Text = append(l.Text, []string{b.String()})
}

func compareAttrandValue(attr html.Attribute, s SearchSpecifc) bool {
	if val, ok := s.Attr[attr.Key]; ok {
		if words := strings.Fields(attr.Val); len(words) > 1 {
			for _, x := range words {
				if x == val {
					return true
				}
			}
		}
		if attr.Val == val {
			return true
		}
	}
	return false
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

func Parse(r io.Reader) (*html.Node, error) {
	return html.Parse(r)
}

func EasyParse(s *[]byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(*s))
}
