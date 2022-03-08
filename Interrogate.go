package speed

import (
	"bytes"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type Attribute struct {
	Name  string
	Value string
}

type NodeList struct {
	Nodes []Node
}

type Node struct {
	Node *html.Node
}

// ----------------------------------------------------------------
func query(r *Node, s Search, l *NodeList) {

	// Search function that carries out checks for attributes
	// appends to node struct to then be appended to another struct called Tag
	// Tag keep all data associated to the to this one struct
	// Search and Tag works on per HTML Doc basis

	var search = func() {
		temp := NodeList{}
		for i := 0; i < len(r.Node.Attr); i++ {
			attr := r.Node.Attr[i]
			if compareQuerySearch(attr, s, r) {
				// temp.append(r.Node)
				temp.Nodes = append(temp.Nodes, *r)
			}
		}

		if len(temp.Nodes) > 0 {
			l.Nodes = append(l.Nodes, temp.Nodes...)
			temp.Nodes = nil
		}
	}

	func() {
		if r.Node.Type == html.ElementNode {
			// checks if a tag exists
			if len(s.Tag) > 0 {
				// check if the tag equals to the data
				if r.Node.Data == s.Tag {
					// checks if the tag is accompanied with attributes/ selectors
					if len(s.Attr) > 0 || len(s.Selector) > 0 {
						search()
					} else { // appends all named tag that has been chosen
						l.Nodes = append(l.Nodes, *r)
					}
				}
			} else { // used if only attr are present for the search
				search()
			}
		}

		// doing a depth first search
		for c := r.Node.FirstChild; c != nil; c = c.NextSibling {
			x := r
			x.Node = c
			query(x, s, l)
		}
	}()

}

func compareQueryAttrList(r *Node, a []Attribute) bool {
	for _, x := range a {
		for _, y := range r.Node.Attr {
			if len(x.Value) > 0 {
				if x.Name == y.Key {
					if ContainsAttrString(y.Val, x.Value) {
						return true
					}
				}
			}

		}
	}

	return false
}

func compareQuerySearch(attr html.Attribute, s Search, r *Node) bool {

	var attrSearch = func(as []Attribute) bool {
		for _, val := range as {
			// if key exists
			if val.Name == attr.Key {
				// if value exists
				if len(val.Value) > 0 {
					// search through attributes
					if ContainsAttrString(attr.Val, val.Value) {
						return true
					}

				} else {
					return true
				}
			}
		}
		return false
	}

	var search = func() bool {

		// if a selector is present
		if len(s.Selector) > 0 {
			// if the selector belong in the document
			if attrSearch(s.Selector) {
				// checks is we need to check for attributes
				if len(s.Attr) > 0 {
					// returns false if the selector and attr don't match
					// we need the node parent for this so we can
					// continue with the search. this is partially broken
					// return attrSearch(s.Attr)
					return compareQueryAttrList(r, s.Attr)
					// return true
				} else {
					return true
				}
			}
			return false
		}

		// for tag + attributes if it has branched off the tag way
		// see if there is attributes that we want to look for
		if len(s.Attr) > 0 {
			return attrSearch(s.Attr)
		}
		return false
	}

	// returns true when only the tag is present
	if len(s.Tag) > 0 && len(s.Attr) == 0 && len(s.Selector) == 0 {
		return true
	}

	return search()
}

// -------------------------------------------------------------------

func searchStrictQueryAttr(attr html.Attribute, val string, num_selector ...int) bool {
	words := strings.Fields(attr.Val)
	// if num_selector != nil {
	// 	if num_selector[0] != len(words) {
	// 		return false
	// 	}
	// }

	if len(words) > 0 {
		for _, x := range words {
			if !(x == val) {
				return false
			}
		}
		return true
	}
	return false
}

func strictQueryCompare(r []html.Attribute, s Search) bool {
	var s_count, a_count int

	s_count = len(s.Selector)
	a_count = len(s.Attr)
	num_selector := s_count

	// fmt.Println("s: ", s_count)
	// fmt.Println("a: ", a_count)

	if s_count > 0 && !(s_count+a_count == len(r)) {
		return false
	}

	// iterating through the nodes attr list
	for _, x := range r {
		// iterating through the selector list we have chosen
		for _, y := range s.Selector {
			// fmt.Println(y)
			// checking if the attr selector exists class/id
			if y.Name == x.Key {
				// if the selector has a value associated with it
				// so the class name/id name
				if len(y.Value) > 0 {
					if searchStrictQueryAttr(x, y.Value, num_selector) {
						s_count--
					}
					// once we have found it we found the key-pair or not we want to continue
					continue
				} else {
					// this is where there is no value associated with the key
					s_count--
				}
			}
		}

		// we going to check if attr list has been given
		if a_count > 0 {
			//  we check if the right length of attributes in html
			// and search attributes/selectors are the same
			if len(r) == len(s.Selector)+len(s.Attr) {
				// loop through the attributes we want to strictly find
				for _, z := range s.Attr {
					// if we found the correct keypair then we increment the counter
					if searchStrictQueryAttr(x, z.Value) {
						a_count--
					}
				}
			}

		}

	}

	//  we are decrementing the count and it should not be higher nor lower than zero
	if s_count == 0 && a_count == 0 {
		return true
	} else {
		return false
	}

}

func strictQueryAttrCompare(r *Node, s Search, l *NodeList) {
	count := len(s.Attr)
	for _, x := range r.Node.Attr {
		for _, y := range s.Attr {
			if y.Name == x.Key {
				if y.Value == x.Val {
					count--
				}
			}
		}
	}
	if count == 0 {
		l.append(r.Node)
	}
}

func strictQuerySearch(r *Node, s Search, l *NodeList) {
	if strictQueryCompare(r.Node.Attr, s) {
		l.append(r.Node)
	}

}

func queryStrictlySearch(r *Node, s Search, l *NodeList) {
	isTag := len(s.Tag) > 0
	isSelector := len(s.Selector) > 0

	// check if it's an element node
	if r.Node.Type == html.ElementNode {
		// check if it equals the tag
		if isTag {
			// current node is equal to our tag?
			if r.Node.Data == s.Tag {
				strictQuerySearch(r, s, l)
			}
		} else if isSelector {
			// runs when only selector present
			strictQuerySearch(r, s, l)
		} else {
			strictQueryAttrCompare(r, s, l)
		}
	}

	for c := r.Node.FirstChild; c != nil; c = c.NextSibling {
		x := r
		x.Node = c
		queryStrictlySearch(x, s, l)
	}

}

func queryStrictly(r *Node, s Search, l *NodeList) {
	queryStrictlySearch(r, s, l)
}

// --------------------------------------------------------------------------------------------
//  HTML parsing. Independeant from the DocumentGroup

func Parse(r io.Reader) (*html.Node, error) {
	return html.Parse(r)
}

func EasyParse(s *[]byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(*s))
}

// ---------------------------------------------------------------
func findAttr(r []html.Attribute, s Search) bool {

	numToBeFound := len(s.Attr)
	for _, x := range r {
		for _, y := range s.Attr {
			if y.Name == x.Key {
				if y.Value == x.Val {
					numToBeFound--
				}
			}
		}
	}

	return numToBeFound == 0
}

func (n *NodeList) append(r *html.Node) {
	n.Nodes = append(n.Nodes, Node{r})
}

func findCompare(r *Node, s Search, l *NodeList) {
	if findAttr(r.Node.Attr, s) {
		l.append(r.Node)
	}
}

func find(r *Node, s Search, l *NodeList) {
	if r.Node.Type == html.ElementNode {

		if len(s.Tag) > 0 {
			if s.Tag == r.Node.Data {
				findCompare(r, s, l)
			}
		} else if len(s.Attr) > 0 {
			findCompare(r, s, l)
		}

	}

	for c := r.Node.FirstChild; c != nil; c = c.NextSibling {
		x := r
		x.Node = c
		find(x, s, l)
	}

}

// func findStrictly(r *Node, s Search, l NodeList){
// 	if r.Node
// }

// --------------------------------------------------------------------------------------------

func ContainsAttrString(str, toCompare string) bool {
	if words := strings.Fields(str); len(words) > 0 {
		// splits atrribute values
		for _, y := range words {
			if toCompare == y {
				return true
			}
		}
	}
	return false
}
