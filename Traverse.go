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

func (h *HTMLDocument) QueryStrictly(search string, m ...func(doc *HTMLDocument)) {
	/*


	 */
	h.NodeList.Nodes = nil
	s := FinderParser(search)

	queryStrictly(h.Node, *s, &h.NodeList)

	if len(m) > 0 {
		m[0](h)
	}

}

func (h *HTMLDocument) Find(search string) *NodeList {
	h.NodeList.Nodes = nil
	s := FinderParser(search)
	find(h.Node, *s, &h.NodeList)
	return &h.NodeList
}

func (h *HTMLDocument) FindStrictly(search string) *NodeList {
	// h.NodeList.Nodes = nil
	s := FinderParser(search)
	findStrictly(h.Node.Node, *s)
	return &h.NodeList
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

func (h *HTMLDocument) PrintNodeList() {
	for _, x := range h.NodeList.Nodes {
		fmt.Println(x)
	}
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

func BreakWords(str string) [][]string {
	var list [][]string
	for _, words := range strings.Fields(str) {
		list = append(list, []string{words})
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

func (n *Node) Attr() map[string]string {

	list := make(map[string]string)
	for _, x := range n.Node.Attr {
		list[x.Key] = x.Val
	}
	return list
}

func (n *Node) Text() string {
	b := &bytes.Buffer{}
	getText(n.Node, b)
	return b.String()
}

func (n *NodeList) GetNode(index int) Node {
	return Node{n.Nodes[index]}
}

func (h *HTMLDocument) Nodify() []Node {
	nodes := make([]Node, 0, 10)
	for _, x := range h.NodeList.Nodes {
		nodes = append(nodes, Node{x})
	}

	h.NodeList.Node = nodes
	return nodes
}
