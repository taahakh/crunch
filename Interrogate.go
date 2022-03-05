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
	Node []Node
}

type Node struct {
	Node *html.Node
}

// Not efficient and takes in Search struct as parameters to automatically search for the correct Nodes
func AdvancedSearch(r *html.Node, s Search, l *Tag) {

	// Search function that carries out checks for attributes
	// appends to node struct to then be appended to another struct called Tag
	// Tag keep all data associated to the to this one struct
	// Search and Tag works on per HTML Doc basis

	var search = func() {
		temp := []Node{}
		for i := 0; i < len(r.Attr); i++ {
			attr := r.Attr[i]
			if compareWithSearch(attr, s, r) {
				// l.Node = append(l.Node, []Node{{Node: r}})
				temp = append(temp, Node{Node: r})
			}
		}

		if len(temp) > 0 {
			l.Node = append(l.Node, temp)
			temp = nil
		}
	}

	func() {
		if r.Type == html.ElementNode {
			// checks if a tag exists
			if len(s.Tag) > 0 {
				// check if the tag equals to the data
				if r.Data == s.Tag {
					// checks if the tag is accompanied with attributes/ selectors
					if len(s.Attr) > 0 || len(s.Selector) > 0 {
						search()
					} else { // appends all named tag that has been chosen
						l.Node = append(l.Node, []Node{{Node: r}})
					}
				}
			} else { // used if only attr are present for the search
				search()
			}
		}

		// doing a depth first search
		for c := r.FirstChild; c != nil; c = c.NextSibling {
			AdvancedSearch(c, s, l)
		}
	}()

	// fmt.Println(sc)
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

func compareWithAttributeList(r *html.Node, a []Attribute) bool {
	for _, x := range a {
		for _, y := range r.Attr {
			if len(x.Value) > 0 {
				if x.Name == y.Key {
					if words := strings.Fields(y.Val); len(words) > 0 {
						for _, z := range words {
							if x.Value == z {
								return true
							}
						}
					}
				}
			}

		}
	}

	return false
}

func compareWithSearch(attr html.Attribute, s Search, r *html.Node) bool {

	var attrSearch = func(as []Attribute) bool {
		for _, val := range as {
			// if key exists
			if val.Name == attr.Key {
				// if value exists
				if len(val.Value) > 0 {
					// search through attributes
					if words := strings.Fields(attr.Val); len(words) > 0 {
						// splits atrribute values
						for _, y := range words {
							if val.Value == y {
								return true
							}
						}
					}
					// return searchAttr(attr, val.Value)
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
					return compareWithAttributeList(r, s.Attr)
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

func searchAttr(attr html.Attribute, val string, num_selector ...int) bool {
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

func strictCompare(r []html.Attribute, s Search) bool {
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
					if searchAttr(x, y.Value, num_selector) {
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
					if searchAttr(x, z.Value) {
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

func strictAttrCompare(r *html.Node, s Search, l *Tag) {
	count := len(s.Attr)
	for _, x := range r.Attr {
		for _, y := range s.Attr {
			if y.Name == x.Key {
				if y.Value == x.Val {
					count--
				}
			}
		}
	}
	if count == 0 {
		l.Node = append(l.Node, []Node{{Node: r}})
	}
}

func strictSearch(r *html.Node, s Search, l *Tag) {
	if strictCompare(r.Attr, s) {
		l.Node = append(l.Node, []Node{{Node: r}})
	}

}

func findStrictlySearch(r *html.Node, s Search, l *Tag) {
	isTag := len(s.Tag) > 0
	isSelector := len(s.Selector) > 0

	// check if it's an element node
	if r.Type == html.ElementNode {
		// check if it equals the tag
		if isTag {
			// current node is equal to our tag?
			if r.Data == s.Tag {
				strictSearch(r, s, l)
			}
		} else if isSelector {
			// runs when only selector present
			strictSearch(r, s, l)
		} else {
			strictAttrCompare(r, s, l)
		}
	}

	for c := r.FirstChild; c != nil; c = c.NextSibling {
		findStrictlySearch(c, s, l)
	}

}

func findStrictly(r *html.Node, s Search, l *Tag) {
	findStrictlySearch(r, s, l)
}

//  HTML parsing. Independeant from the DocumentGroup

func Parse(r io.Reader) (*html.Node, error) {
	return html.Parse(r)
}

func EasyParse(s *[]byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(*s))
}

func qS(r []html.Attribute, s Search) bool {

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

// func qSCompare(r *html.Node, s Search, l *NodeList) {
// 	if qS(r.Attr, s) {
// 		l.Node = append(l.Node, r)
// 	}
// }

// func querySearch(r *html.Node, s Search, l *NodeList) {
// 	if r.Type == html.ElementNode {

// 		if len(s.Tag) > 0 {
// 			if s.Tag == r.Data {
// 				qSCompare(r, s, l)
// 			}
// 		} else if len(s.Attr) > 0 {
// 			qSCompare(r, s, l)
// 		}

// 	}
// 	for c := r.FirstChild; c != nil; c = c.NextSibling {
// 		querySearch(c, s, l)
// 	}
// }

// func qS(r []html.Attribute, s Search) bool {

// 	numToBeFound := len(s.Attr)
// 	for _, x := range r {
// 		for _, y := range s.Attr {
// 			if y.Name == x.Key {
// 				if y.Value == x.Val {
// 					numToBeFound--
// 				}
// 			}
// 		}
// 	}

// 	return numToBeFound == 0
// }
func (n *NodeList) append(r *html.Node) {
	n.Node = append(n.Node, Node{r})
}

func qSCompare(r *Node, s Search, l *NodeList) {
	if qS(r.Node.Attr, s) {
		l.append(r.Node)
	}
}

func querySearch(r *Node, s Search, l *NodeList) {
	// fmt.Println(r)
	if r.Node.Type == html.ElementNode {

		if len(s.Tag) > 0 {
			if s.Tag == r.Node.Data {
				qSCompare(r, s, l)
			}
		} else if len(s.Attr) > 0 {
			qSCompare(r, s, l)
		}

	}

	for c := r.Node.FirstChild; c != nil; c = c.NextSibling {
		x := r
		x.Node = c
		querySearch(x, s, l)
	}

}

func (n *Node) Text() string {
	b := &bytes.Buffer{}
	getText(n.Node, b)
	return b.String()
}
