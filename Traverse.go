package speed

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/net/html"
)

var (
	f func(x *html.Node)
)

// Node is the HTML document
// Nodelist are the pointer list to the nodes that the tag/selector has been searched for
// Simple struct to store relevant search data for the document
// NodeList will get wiped for each search. Searched Data can be saved after search is their is a return type (usually []/*html.Node)
type HTMLDocument struct {
	Node     *html.Node
	NodeList []*html.Node
}

type DocumentGroup struct {
	Collector []HTMLDocument
}

// Add to the doc struct for each HTML
func CreateHTMLDocument(r io.Reader) HTMLDocument {
	doc, err := html.Parse(r)
	if err != nil {
		fmt.Println("something")
	}
	return HTMLDocument{Node: doc}
}

func (h *HTMLDocument) FindTag(element string, m ...func(doc *HTMLDocument)) []*html.Node {
	f = func(x *html.Node) {
		if x.Type == html.ElementNode && x.Data == element {
			h.NodeList = append(h.NodeList, x)
		}

		for c := x.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(h.Node)
	for _, x := range m {
		x(h)
	}
	return h.NodeList
}

func (h *HTMLDocument) Find(search string, m ...func(doc *HTMLDocument)) {
	s := FinderParser(search)
	l := &Tag{}
	// fmt.Println(FinderParser(search))
	AdvancedSearch(h.Node, *s, l)
	for _, x := range l.Node {
		for _, y := range x {
			fmt.Println(y.Node)
		}
	}
}

func (h *HTMLDocument) Attr() []map[string]string {
	list := make([]map[string]string, 0)
	for _, x := range h.NodeList {
		t := make(map[string]string, 0)
		for _, y := range x.Attr {
			t[y.Key] = y.Val
		}
		list = append(list, t)
	}

	return list
}

func (h *HTMLDocument) GetAttr(elem ...string) []string {
	return getAttr(h.NodeList, false, elem)
}

func (h *HTMLDocument) GetAttrOnce(elem ...string) string {
	return getAttr(h.NodeList, true, elem)[0]
}

func getAttr(r []*html.Node, once bool, elem []string) []string {
	var list []string
	if once {
		list = make([]string, 0)
	}
	for _, x := range r {
		for _, y := range x.Attr {
			for _, z := range elem {
				if y.Key == z {
					if once {
						return []string{y.Val}
					}
					list = append(list, y.Val)
				}
			}
		}
	}
	return list
}

func Text(r *html.Node, b *bytes.Buffer) string {
	if r.Type == html.TextNode {
		b.WriteString(r.Data)
	}
	return b.String()
}

// func Find(h *html.Node, element string) []*html.Node {
// 	list := make([]*html.Node, 0)

// 	var f func(x *html.Node)
// 	f = func(x *html.Node) {
// 		if x.Type == html.ElementNode && x.Data == element {
// 			list = append(list, x)
// 		}

// 		for c := x.FirstChild; c != nil; c = c.NextSibling {
// 			f(c)
// 		}
// 	}
// 	f(h)
// 	return list
// }
