package traverse

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// Node is the HTML document
// Nodelist are the pointer list to the nodes that the tag/selector has been searched for
// NodeList will get wiped for each search. Searched Data can be saved after search is their is a return type (usually []/*html.Node)
// InitialSearch and Complete tracks how the Doc will be traversed
// InitialSearch controls what nodes to apply functions
// Complete dictates where to branch off to a new doc node
type HTMLDocument struct {
	Node         Node     //HTML DOC Node
	NodeList     NodeList // Current search result
	IntialSearch bool
	Complete     bool // All searches and data interpretation is done on one object
}

// ----------------------------------------------------------------------------------------------------------

// Add to the doc struct for each HTML
// Decided to put Complete to true straight away
// More convenient to code this way
func HTMLDoc(r io.Reader) (*HTMLDocument, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, errors.New("There was an error parsing")
	}
	return &HTMLDocument{Node: Node{Node: doc}, IntialSearch: false, Complete: true}, nil
}

func HTMLDocRun(r io.Reader, m func(doc *HTMLDocument) bool) bool {
	doc, err := html.Parse(r)
	if err != nil {
		fmt.Println("something")
	}

	document := HTMLDocument{Node: Node{Node: doc}, IntialSearch: false, Complete: true}

	return m(&document)
}

func HTMLDocBytes(b *[]byte) HTMLDocument {
	doc, err := StringEasyParse(b)
	if err != nil {
		fmt.Println("something")
	}
	return HTMLDocument{Node: Node{Node: doc}, IntialSearch: false, Complete: true}
}

// Requests are going to be encoded in utf-8 so it needs to be set in a format readable by the parser
func HTMLDocUTF8(r *http.Response) (HTMLDocument, error) {
	defer r.Body.Close()
	utf8set, err := charset.NewReader(r.Body, r.Header.Get("Content-Type"))
	if err != nil {
		log.Println("Failed utf8set")
	}
	bytes, err := ioutil.ReadAll(utf8set)
	if err != nil {
		log.Println("Failed ioutil")
	}
	return HTMLDocBytes(&bytes), err
}

func HTMLNodeToDoc(n *html.Node) HTMLDocument {
	return HTMLDocument{Node: Node{Node: n}, IntialSearch: false, Complete: true}
}

// ----------------------------------------------------------------------------------------------------------

// Selects control if they are going to search once or until all found
// Controls how data is going to be stored and returned
func (h *HTMLDocument) QuerySelect(search string, once bool) *HTMLDocument {
	var tempAppend []*html.Node
	s := FinderParser(search)
	if !(h.IntialSearch) {
		if !(h.Complete) {
			h.IntialSearch = true
		}
		tempAppend = query(h.Node.Node, *s, once)

	} else {
		if once {
			for _, x := range h.NodeList {
				tempAppend = query(x, *s, true)
				if len(tempAppend) == 1 {
					break
				}
			}

		} else {
			for _, x := range h.NodeList {
				temp := query(x, *s, false)
				tempAppend = append(tempAppend, temp...)
			}
		}

	}

	if h.Complete {
		return &HTMLDocument{Node: h.Node, NodeList: tempAppend, IntialSearch: true}
	}

	h.NodeList = tempAppend

	return h
}

func (h *HTMLDocument) FindSelect(search string, once bool) *HTMLDocument {
	var tempAppend []*html.Node
	s := FinderParser(search)
	if !(h.IntialSearch) {
		if !(h.Complete) {
			h.IntialSearch = true
		}
		tempAppend = find(h.Node.Node, *s, once)
	} else {

		if once {
			for _, x := range h.NodeList {
				tempAppend = find(x, *s, true)
				if len(tempAppend) == 1 {
					break
				}
			}
		} else {
			for _, x := range h.NodeList {
				temp := find(x, *s, false)
				tempAppend = append(tempAppend, temp...)
			}
		}

	}

	if h.Complete {
		return &HTMLDocument{Node: h.Node, NodeList: tempAppend, IntialSearch: true}
	}

	h.NodeList = tempAppend

	return h
}

func (h *HTMLDocument) FindStrictlySelect(search string, once bool) *HTMLDocument {
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
		tempAppend = findStrictly(h.Node.Node, *s, once)
	} else {
		if len(s.Attr) == 0 && len(s.Tag) > 0 || len(s.Selector) > 0 {
			return h
		}
		if once {
			for _, x := range h.NodeList {
				tempAppend = findStrictly(x, *s, true)
				if len(tempAppend) == 1 {
					break
				}
			}
		} else {
			for _, x := range h.NodeList {
				temp := findStrictly(x, *s, false)
				tempAppend = append(tempAppend, temp...)
			}

		}

	}

	if h.Complete {
		return &HTMLDocument{Node: h.Node, NodeList: tempAppend, IntialSearch: true}
	}

	h.NodeList = tempAppend

	return h
}

// ----------------------------------------------------------------------------------------------------------

// Simple function that calls any of the traversal code
// More convenient and code in one place
// Individual implementation are available and should be the main way to traverse the DOM
// Can call a function to make code neater
func (h *HTMLDocument) Search(f string, search string, m ...func(doc *HTMLDocument)) *HTMLDocument {
	// var doc *HTMLDocument
	// switch f {
	// case "query":
	// 	doc = h.QuerySelect(search, false)
	// 	break
	// case "queryOnce":
	// 	doc = h.QuerySelect(search, true)
	// 	break
	// case "find":
	// 	doc = h.FindSelect(search, false)
	// 	break
	// case "findOnce":
	// 	doc = h.FindSelect(search, true)
	// 	break
	// case "findStrict":
	// 	doc = h.FindStrictlySelect(search, false)
	// 	break
	// case "findStrictOnce":
	// 	doc = h.FindStrictlySelect(search, true)
	// 	break
	// default:
	// 	doc = h
	// 	break
	// }

	SearchSwitch(f, search, h)

	if m != nil {
		m[0](h)
	}

	return h
}

func SearchSwitch(function, search string, doc *HTMLDocument) {
	switch function {
	case "query":
		doc.QuerySelect(search, false)
		break
	case "queryOnce":
		doc.QuerySelect(search, true)
		break
	case "find":
		doc.FindSelect(search, false)
		break
	case "findOnce":
		doc.FindSelect(search, true)
		break
	case "findStrict":
		doc.FindStrictlySelect(search, false)
		break
	case "findStrictOnce":
		doc.FindStrictlySelect(search, true)
		break
	default:
		break
	}
}

func (h *HTMLDocument) Query(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {

	doc := h.QuerySelect(search, false)

	// if m != nil {
	// 	m[0](doc)
	// }

	runCustom(doc, m...)

	return doc
}

func (h *HTMLDocument) QueryOnce(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {
	doc := h.QuerySelect(search, true)

	// if m != nil {
	// 	m[0](doc)
	// }

	runCustom(doc, m...)

	return doc
}

func (h *HTMLDocument) Find(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {

	doc := h.FindSelect(search, false)

	// if m != nil {
	// 	m[0](doc)
	// }

	runCustom(doc, m...)

	return doc
}

func (h *HTMLDocument) FindOnce(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {
	doc := h.FindSelect(search, true)

	// if m != nil {
	// 	m[0](doc)
	// }

	runCustom(doc, m...)

	return doc
}

func (h *HTMLDocument) FindStrictly(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {
	doc := h.FindStrictlySelect(search, false)

	// if m != nil {
	// 	m[0](doc)
	// }

	runCustom(doc, m...)

	return doc
}

func (h *HTMLDocument) FindStrictlyOnce(search string, m ...func(doc *HTMLDocument)) *HTMLDocument {
	doc := h.FindStrictlySelect(search, true)

	// if m != nil {
	// 	m[0](doc)
	// }

	runCustom(doc, m...)

	return doc
}

func runCustom(item *HTMLDocument, m ...func(doc *HTMLDocument)) {
	if m != nil {
		m[0](item)
	}
}

// ----------------------------------------------------------------------------------------------------------

// Creates an attribute map. All nodes attributes and values are gathered - put together
func (h *HTMLDocument) Attr() map[string][]string {

	list := make(map[string][]string, 0)
	for _, x := range h.NodeList {
		for _, y := range x.Attr {
			if count, ok := list[y.Key]; ok {
				list[y.Key] = append(count, y.Val)
			} else {
				list[y.Key] = []string{y.Val}
			}
		}
	}

	return list
}

// Specifically ask for the attr(s) that you want
func (h *HTMLDocument) GetAttr(elem ...string) []string {
	return getAttr(h.NodeList, false, elem)
}

// Returns only the first attribute
func (h *HTMLDocument) GetAttrOnce(elem ...string) string {
	return getAttr(h.NodeList, true, elem)[0]
}

// Prints current node list
func (h *HTMLDocument) PrintNodeList() {
	for _, x := range h.NodeList {
		fmt.Println(x)
	}
}

// Makes Grouped searches
// When the document is set to Done. It will return a new copy of the document
// Traversal/Searches can be made from that point on without affecting the original document
func (h *HTMLDocument) Done() {
	h.Complete = true
}

// Resets the document search by starting from another node
func (h *HTMLDocument) SetNode(i int) *HTMLDocument {
	h.Node = h.NodeList.GetNode(i)
	h.NodeList = nil
	h.IntialSearch = false
	return h
}

// Return given node
func (h *HTMLDocument) GetNode(i int) Node {
	return h.NodeList.GetNode(i)
}

// Returns current nodelist
func (h *HTMLDocument) GiveNodeList() NodeList {
	return h.NodeList
}

// Makes all nodes in NodeList part of Node{}
// Should be avoided
func (h *HTMLDocument) Nodify() []Node {
	nodes := make([]Node, len(h.NodeList))
	for _, x := range h.NodeList {
		nodes = append(nodes, Node{x})
	}

	return nodes
}

// We can iterate through the nodes through a single function
// Some what similar to Nodify but Nodes are created on the go
// and actions are taken
func (h *HTMLDocument) Iterate(m func(doc Node)) {
	for _, x := range h.NodeList {
		m(Node{x})
	}
}

// Iterates until its finished
func (h *HTMLDocument) IterateBreak(m func(doc Node) bool) {
	for _, x := range h.NodeList {
		b := m(Node{x})
		if b {
			break
		}
	}
}

// Returns nodes attr
func (n *Node) Attr() map[string]string {

	list := make(map[string]string)
	for _, x := range n.Node.Attr {
		list[x.Key] = x.Val
	}
	return list
}

// DOM node traversal
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

	return &HTMLDocument{Node: *n, NodeList: nodes, IntialSearch: true}
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
	return &HTMLDocument{Node: *n, NodeList: nodes, IntialSearch: true}
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

// Removes any whitespace.
func (n Node) CleanText() string {
	return strings.TrimSpace(n.Text())
}

// Renders the HTML of that node
func (n *Node) RenderNode() string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n.Node)
	return buf.String()
}

// Simple attr finder
func (n *Node) FindAttr(attr string) string {
	for _, x := range n.Node.Attr {
		if x.Key == attr {
			return x.Val
		}
	}
	return ""
}

func (n NodeList) GetNode(index int) Node {
	return Node{n[index]}
}

func getAttr(r NodeList, once bool, elem []string) []string {
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

// Wraps *html.Node to Node
func ToNode(r *html.Node) *Node {
	return &Node{r}
}
