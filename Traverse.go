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
	Node         Node     //HTML DOC Node
	NodeList     NodeList // Current search result
	IntialSearch bool
	Complete     bool
}

type DocumentGroup struct {
	Collector []HTMLDocument
}

// Add to the doc struct for each HTML
func HTMLDoc(r io.Reader) HTMLDocument {
	doc, err := html.Parse(r)
	if err != nil {
		fmt.Println("something")
	}
	return HTMLDocument{Node: Node{Node: doc}, IntialSearch: false, Complete: true}
}

func (h *HTMLDocument) querySelect(search string, once bool) *HTMLDocument {
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
	var tempAppend []*html.Node
	s := FinderParser(search)
	if !(h.IntialSearch) {
		// if !(h.Simple) {
		if !(h.Complete) {
			h.IntialSearch = true
		}
		// }
		// h.NodeList.Nodes = query(h.Node.Node, *s, once)
		tempAppend = query(h.Node.Node, *s, once)

	} else {
		if once {
			for _, x := range h.NodeList.Nodes {
				tempAppend = query(x, *s, true)
				if len(tempAppend) == 1 {
					break
				}
			}

		} else {
			for _, x := range h.NodeList.Nodes {
				temp := query(x, *s, false)
				tempAppend = append(tempAppend, temp...)
			}
		}

	}

	if h.Complete {
		return &HTMLDocument{Node: h.Node, NodeList: NodeList{Nodes: tempAppend}, IntialSearch: true}
	}

	h.NodeList.Nodes = tempAppend

	return h
}

func (h *HTMLDocument) findSelect(search string, once bool) *HTMLDocument {
	var tempAppend []*html.Node
	s := FinderParser(search)
	if !(h.IntialSearch) {
		if !(h.Complete) {
			h.IntialSearch = true
		}
		tempAppend = find(h.Node.Node, *s, once)
		// h.NodeList.Nodes = find(h.Node.Node, *s, once)
	} else {

		if once {
			for _, x := range h.NodeList.Nodes {
				tempAppend = find(x, *s, true)
				if len(tempAppend) == 1 {
					break
				}
			}
		} else {
			for _, x := range h.NodeList.Nodes {
				temp := find(x, *s, false)
				tempAppend = append(tempAppend, temp...)
			}
		}

	}

	if h.Complete {
		return &HTMLDocument{Node: h.Node, NodeList: NodeList{Nodes: tempAppend}, IntialSearch: true}
	}

	h.NodeList.Nodes = tempAppend

	return h
}

func (h *HTMLDocument) findStrictlySelect(search string, once bool) *HTMLDocument {
	var tempAppend []*html.Node
	s := FinderParser(search)
	if !(h.IntialSearch) {
		if !(h.Complete) {
			h.IntialSearch = true
		}
		// we don't want to place a tag alone
		// left side selector does not work
		if len(s.Attr) == 0 && len(s.Tag) > 0 || len(s.Selector) > 0 {
			return h
		}
		// h.NodeList.Nodes = findStrictly(h.Node.Node, *s, once)
		tempAppend = findStrictly(h.Node.Node, *s, once)
	} else {
		if len(s.Attr) == 0 && len(s.Tag) > 0 || len(s.Selector) > 0 {
			return h
		}
		if once {
			for _, x := range h.NodeList.Nodes {
				tempAppend = findStrictly(x, *s, true)
				if len(tempAppend) == 1 {
					break
				}
			}
		} else {
			for _, x := range h.NodeList.Nodes {
				temp := findStrictly(x, *s, false)
				tempAppend = append(tempAppend, temp...)
			}

		}

	}

	if h.Complete {
		return &HTMLDocument{Node: h.Node, NodeList: NodeList{Nodes: tempAppend}, IntialSearch: true}
	}

	h.NodeList.Nodes = tempAppend

	return h
}

func (h *HTMLDocument) Search(f string, search string, m ...func(doc *HTMLDocument)) *HTMLDocument {
	var doc *HTMLDocument
	switch f {
	case "query":
		doc = h.querySelect(search, false)
		break
	case "queryOnce":
		doc = h.querySelect(search, true)
		break
	case "find":
		doc = h.findSelect(search, false)
		break
	case "findOnce":
		doc = h.findSelect(search, true)
		break
	case "findStrict":
		doc = h.findStrictlySelect(search, false)
		break
	case "findStrictOnce":
		doc = h.findStrictlySelect(search, true)
		break
	default:
		doc = h
		break
	}

	if m != nil {
		m[0](doc)
	}

	return doc
}

func (h *HTMLDocument) Query(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {

	doc := h.querySelect(search, false)

	if m != nil {
		m[0](doc)
	}

	return doc
}

func (h *HTMLDocument) QueryOnce(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {
	doc := h.querySelect(search, true)

	if m != nil {
		m[0](doc)
	}

	return doc
}

func (h *HTMLDocument) Find(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {

	doc := h.findSelect(search, false)

	if m != nil {
		m[0](doc)
	}

	return doc
}

func (h *HTMLDocument) FindOnce(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {
	doc := h.findSelect(search, true)

	if m != nil {
		m[0](doc)
	}

	return doc
}

func (h *HTMLDocument) FindStrictly(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {
	doc := h.findStrictlySelect(search, false)

	if m != nil {
		m[0](doc)
	}

	return doc
}

func (h *HTMLDocument) FindStrictlyOnce(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {
	doc := h.findStrictlySelect(search, true)

	if m != nil {
		m[0](doc)
	}

	return doc
}

func (h *HTMLDocument) Attr() map[string][]string {

	list := make(map[string][]string, 0)
	for _, x := range h.NodeList.Nodes {
		// t := make(map[string]string, 0)
		for _, y := range x.Attr {
			// t[y.Key] = y.Val
			if count, ok := list[y.Key]; ok {
				list[y.Key] = append(count, y.Val)
			} else {
				list[y.Key] = []string{y.Val}
			}
		}
		// list = append(list, t)
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

func (h *HTMLDocument) Done() {
	h.Complete = true
}

func (h *HTMLDocument) SetNode(i int) *HTMLDocument {
	h.Node = h.NodeList.GetNode(i)
	h.NodeList.Nodes = nil
	h.IntialSearch = false
	return h
}

func (h *HTMLDocument) GetNode(i int) Node {
	return h.NodeList.GetNode(i)
}

func (h *HTMLDocument) GiveNodeList() NodeList {
	return h.NodeList
}

func (h *HTMLDocument) GiveHTMLNodes() []*html.Node {
	return h.NodeList.Nodes
}

func (h *HTMLDocument) Nodify() []Node {
	nodes := make([]Node, 0, 10)
	for _, x := range h.NodeList.Nodes {
		nodes = append(nodes, Node{x})
	}

	h.NodeList.nodes = nodes
	return nodes
}

func (h *HTMLDocument) Iterate(m func(doc Node)) {
	for _, x := range h.NodeList.Nodes {
		m(Node{x})
	}
}

func (h *HTMLDocument) IterateBreak(m func(doc Node) bool) {
	for _, x := range h.NodeList.Nodes {
		b := m(Node{x})
		if b {
			break
		}
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

func (n *Node) ChildrenNode() *HTMLDocument {
	var f func(r *html.Node)
	nodes := make([]*html.Node, 0, 10)

	f = func(r *html.Node) {
		if r.Type == html.ElementNode {
			nodes = append(nodes, r)
		}

		for c := r.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n.Node)

	return &HTMLDocument{Node: *n, NodeList: NodeList{Nodes: nodes}, IntialSearch: true}
}

func (n *Node) DirectChildrenNode() *HTMLDocument {
	var f func(r *html.Node)
	nodes := make([]*html.Node, 0, 10)

	f = func(r *html.Node) {

		c := r.NextSibling
		for {
			if c.Type == html.ElementNode {
				nodes = append(nodes, c)
			}
			c = c.NextSibling
			if c == nil {
				break
			}
		}
	}

	f(n.Node.FirstChild)
	return &HTMLDocument{Node: *n, NodeList: NodeList{Nodes: nodes}, IntialSearch: true}
}

func (n Node) Text() string {
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

func (n Node) CleanText() string {
	return strings.TrimSpace(n.Text())
}

func (n *Node) RenderNode() string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n.Node)
	return buf.String()
}

func (n *Node) FindAttr(attr string) string {
	for _, x := range n.Node.Attr {
		if x.Key == attr {
			return x.Val
		}
	}
	return ""
}

func (n *NodeList) GetNode(index int) Node {
	return Node{n.Nodes[index]}
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
