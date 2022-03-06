package speed

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/net/html"
)

// Node is the HTML document
// Nodelist are the pointer list to the nodes that the tag/selector has been searched for
// Simple struct to store relevant search data for the document
// NodeList will get wiped for each search. Searched Data can be saved after search is their is a return type (usually []/*html.Node)

type HTMLDocument struct {
	Node     *Node    //HTML DOC Node
	NodeList NodeList // Current search result
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
	return HTMLDocument{Node: &Node{Node: doc}}
}

func (h *HTMLDocument) FindTag(element string, m ...func(doc *HTMLDocument)) []*html.Node {
	// f = func(x *html.Node) {
	// 	if x.Type == html.ElementNode && x.Data == element {
	// 		h.NodeList = append(h.NodeList, x)
	// 	}

	// 	for c := x.FirstChild; c != nil; c = c.NextSibling {
	// 		f(c)
	// 	}
	// }
	// f(h.Node)
	// for _, x := range m {
	// 	x(h)
	// }
	// return h.NodeList

	fmt.Println(element)
	return []*html.Node{}
}

func (h *HTMLDocument) Find(search string, m ...func(doc *HTMLDocument)) *NodeList {
	s := FinderParser(search)
	AdvancedSearch(h.Node, *s, &h.NodeList)
	return &h.NodeList
}

func (h *HTMLDocument) FindStrictly(search string, m ...func(doc *HTMLDocument)) {
	s := FinderParser(search)
	findStrictly(h.Node, *s, &h.NodeList)

	if len(m) > 0 {
		m[0](h)
	}

}

func (h *HTMLDocument) QuerySearch(search string) *NodeList {
	h.NodeList.Nodes = nil
	s := FinderParser(search)
	querySearch(h.Node, *s, &h.NodeList)
	return &h.NodeList
}

func (h *HTMLDocument) Attr() []map[string]string {
	list := make([]map[string]string, 0)
	for _, x := range h.NodeList.Nodes {
		t := make(map[string]string, 0)
		for _, y := range x.Node.Attr {
			t[y.Key] = y.Val
		}
		list = append(list, t)
	}

	return list
}

func (n *Node) Attr() map[string]string {
	defer Catch_Panic()
	list := make(map[string]string)
	for _, x := range n.Node.Attr {
		list[x.Key] = x.Val
	}
	return list
}

func (h *HTMLDocument) GetAttr(elem ...string) []string {
	return getAttr(h.NodeList, false, elem)
}

func (h *HTMLDocument) GetAttrOnce(elem ...string) string {
	return getAttr(h.NodeList, true, elem)[0]
}

func getAttr(r NodeList, once bool, elem []string) []string {
	var list []string
	if once {
		list = make([]string, 0)
	}
	for _, x := range r.Nodes {
		for _, y := range x.Node.Attr {
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

func (h *HTMLDocument) PrintNodeList() {
	for _, x := range h.NodeList.Nodes {
		fmt.Println(x.Node)
	}
}

func (n *Node) Text() string {
	b := &bytes.Buffer{}
	getText(n.Node, b)
	return b.String()
}

func Text(r *html.Node) string {
	b := &bytes.Buffer{}
	getText(r, b)
	return b.String()
}

func getText(r *html.Node, b *bytes.Buffer) {
	if r.Type == html.TextNode {
		b.WriteString(r.Data)
	}
	for c := r.FirstChild; c != nil; c = c.NextSibling {
		getText(c, b)
	}
}
