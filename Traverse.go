package speed

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Node is the HTML document
// Nodelist are the pointer list to the nodes that the tag/selector has been searched for
// Simple struct to store relevant search data for the document
// NodeList will get wiped for each search. Searched Data can be saved after search is their is a return type (usually []/*html.Node)

type HTMLDocument struct {
	Node         *Node    //HTML DOC Node
	NodeList     NodeList // Current search result
	IntialSearch bool
}

type DocumentGroup struct {
	Collector []HTMLDocument
}

type SearchingTypes interface {
}

// Add to the doc struct for each HTML
func CreateHTMLDocument(r io.Reader) HTMLDocument {
	doc, err := html.Parse(r)
	if err != nil {
		fmt.Println("something")
	}
	return HTMLDocument{Node: &Node{Node: doc}, IntialSearch: false}
}

func (h *HTMLDocument) Query(search string, m ...func(doc *HTMLDocument)) *NodeList {
	/*
		SHOULD BE USED FOR SMALL NUMBER OF NODES AND FOR FINDING TAGS

		tag/tag(. or #)selector --> strict
		[] --> non-strict, key-independant, key-pair attr
		--> finds all nodes where the string is present.
			e.g. .box will find nodes that contains this word in the class even if it has multiple classes
		--> attr box [] can search for nodes that the attr you want
			NOTE: Most tags for searching data will usually not have anything other than div or id
				  This is for tags such as <link>
			link[crossorigin]
			link[href='style.css']
			link[crossorigin, href='style.css']

			The search is done loosely in the attr box []. It will return nodes that have the attrs but doesn't
			mean it will strictly follow a rule to return tags that only contains those attrs

		Max accepted string
			tag.selector[attr='', attr='']

	*/
	h.NodeList.Nodes = nil
	s := FinderParser(search)

	query(h.Node, *s, &h.NodeList)
	return &h.NodeList
}

func (h *HTMLDocument) NewQuery(search string) *HTMLDocument {
	h.NodeList.Nodes = nil
	s := FinderParser(search)
	h.NodeList.Nodes = newQuery(h.Node.Node, *s)
	return h
}

func (h *HTMLDocument) QueryStrictly(search string, m ...func(doc *HTMLDocument)) {
	/*


	 */
	h.NodeList.Nodes = nil
	s := FinderParser(search)

	queryStrictly(h.Node.Node, *s, &h.NodeList, false)

	if len(m) > 0 {
		m[0](h)
	}
}

func (h *HTMLDocument) QueryStrictlyOnce(search string, m ...func(doc *HTMLDocument)) {
	h.NodeList.Nodes = nil
	s := FinderParser(search)

	queryStrictly(h.Node.Node, *s, &h.NodeList, true)

	if len(m) > 0 {
		m[0](h)
	}
}

func (h *HTMLDocument) Find(search string) *HTMLDocument {
	s := FinderParser(search)
	if !(h.IntialSearch) {
		h.NodeList.Nodes = find(h.Node.Node, *s, false)
		h.IntialSearch = true
	} else {
		tempAppend := make([]*html.Node, 0, 10)
		for _, x := range h.NodeList.Nodes {
			temp := find(x, *s, false)
			tempAppend = append(tempAppend, temp...)
		}

		h.NodeList.Nodes = tempAppend

	}
	return h
}

func (h *HTMLDocument) FindOnce(search string) *HTMLDocument {
	s := FinderParser(search)
	var temp []*html.Node
	if !(h.IntialSearch) {
		h.NodeList.Nodes = find(h.Node.Node, *s, true)
		// if len(h.NodeList.Nodes) != 0 {
		// 	h.Node = ToNode(h.NodeList.Nodes[0])
		// }
		h.IntialSearch = true
	} else {
		for _, x := range h.NodeList.Nodes {
			temp = find(x, *s, true)
			if len(temp) == 1 {
				break
			}
		}
		h.NodeList.Nodes = temp
	}

	return h
}

func (h *HTMLDocument) FindStrictly(search string) *HTMLDocument {
	s := FinderParser(search)
	if !(h.IntialSearch) {
		h.IntialSearch = true
		// we don't want to place a tag alone
		// left side selector does not work
		if len(s.Attr) == 0 && len(s.Tag) > 0 || len(s.Selector) > 0 {
			return h
		}
		h.NodeList.Nodes = findStrictly(h.Node.Node, *s, false)
	} else {
		if len(s.Attr) == 0 && len(s.Tag) > 0 || len(s.Selector) > 0 {
			return h
		}
		tempAppend := make([]*html.Node, 0, 10)
		for _, x := range h.NodeList.Nodes {
			temp := findStrictly(x, *s, false)
			tempAppend = append(tempAppend, temp...)
		}

		h.NodeList.Nodes = tempAppend
	}

	return h
}

func (h *HTMLDocument) FindStrictlyOnce(search string) *HTMLDocument {
	s := FinderParser(search)
	var temp []*html.Node
	if !(h.IntialSearch) {
		h.IntialSearch = true
		if len(s.Attr) == 0 && len(s.Tag) > 0 || len(s.Selector) > 0 {
			return h
		}
		h.NodeList.Nodes = findStrictly(h.Node.Node, *s, true)
	} else {
		for _, x := range h.NodeList.Nodes {
			temp = findStrictly(x, *s, true)
			if len(temp) == 1 {
				break
			}
		}
		h.NodeList.Nodes = temp
	}

	return h
}

func (h *HTMLDocument) Attr() []map[string]string {

	list := make([]map[string]string, 0)
	for _, x := range h.NodeList.Nodes {
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
	defer Catch_Panic()
	return getAttr(h.NodeList, true, elem)[0]
}

func (h *HTMLDocument) PrintNodeList() {
	for _, x := range h.NodeList.Nodes {
		fmt.Println(x)
	}
}

func (n *Node) Attr() map[string]string {

	list := make(map[string]string)
	for _, x := range n.Node.Attr {
		list[x.Key] = x.Val
	}
	return list
}

func (n *Node) PrevSiblingNode() Node {
	if n.Node.PrevSibling != nil {
		n.Node = n.Node.PrevSibling
		return Node{n.Node}
	}
	return Node{}
}

func (n *Node) PrevElementNode() Node {
	if n.Node.PrevSibling != nil {
		if n.Node.PrevSibling.Type == html.ElementNode {
			n.Node = n.Node.PrevSibling
			return Node{n.Node}
		}
	}
	return Node{}
}

func (n *Node) NextSiblingNode() Node {
	if n.Node.NextSibling != nil {
		n.Node = n.Node.NextSibling
		return Node{n.Node}
	}
	return Node{}
}

func (n *Node) NextElementNode() Node {
	if n.Node.NextSibling != nil {
		if n.Node.NextSibling.Type == html.ElementNode {
			n.Node = n.Node.NextSibling
			return Node{n.Node}
		}
	}
	return Node{}
}

func (n *Node) ChildrenNode() []Node {
	var f func(r *html.Node)
	nodes := make([]Node, 0, 10)

	f = func(r *html.Node) {
		if r.Type == html.ElementNode {
			nodes = append(nodes, Node{r})
		}

		for c := r.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n.Node)
	return nodes
}

func (n *Node) Text() string {
	b := &bytes.Buffer{}
	var f func(r *html.Node)
	f = func(r *html.Node) {
		if r.Type == html.TextNode {
			b.WriteString(r.Data)
		}
		for c := r.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n.Node)

	return b.String()
}

func (n *Node) RenderNode() string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n.Node)
	return buf.String()
}

func (n *NodeList) GetNode(index int) Node {
	return Node{n.Nodes[index]}
}

func (h *HTMLDocument) Nodify() []Node {
	nodes := make([]Node, 0, 10)
	for _, x := range h.NodeList.Nodes {
		nodes = append(nodes, Node{x})
	}

	h.NodeList.nodes = nodes
	return nodes
}

func (h *HTMLDocument) Docify() []HTMLDocument {
	doc := make([]HTMLDocument, 0, 3)
	for _, x := range h.NodeList.Nodes {
		doc = append(doc, HTMLDocument{Node: &Node{x}})
	}
	return doc
}

func getAttr(r NodeList, once bool, elem []string) []string {
	var list []string
	if once {
		list = make([]string, 0)
	}
	for _, x := range r.Nodes {
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

func BreakWords(str string) []string {
	var list []string
	list = append(list, strings.Fields(str)...)
	return list
}

func Exetime(name string) func() {
	start := time.Now()
	return func() {
		x := time.Since(start)
		log.Printf("%s, execution time %s\n", name, x)
		log.Println(x.Microseconds())
	}
}

func ToNode(r *html.Node) *Node {
	return &Node{r}
}

func CompareAttrLists(search string, node string) bool {
	searchList := BreakWords(search)
	nodeList := BreakWords(node)
	min := len(searchList)

	for _, x := range searchList {
		for _, y := range nodeList {
			if x == y {
				min--
			}
		}
	}

	return min == 0
}
