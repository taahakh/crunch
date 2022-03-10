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
	Nodes []*html.Node // lets users have their freedom to manipulate the raw html.Node
	nodes []Node       // private as we want methods to dictate what to do with these nodes
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
	var f func(r *html.Node, s Search)
	nodes := make([]*html.Node, 0, 10)

	var search = func(r *html.Node) {
		temp := make([]*html.Node, 0)
		for i := 0; i < len(r.Attr); i++ {
			attr := r.Attr[i]
			if compareQuerySearch(attr, s, r) {
				temp = append(temp, r)
			}
		}

		if len(temp) > 0 {
			nodes = append(nodes, temp...)
			temp = nil
		}

	}

	f = func(r *html.Node, s Search) {
		if r.Type == html.ElementNode {
			// checks if a tag exists
			if len(s.Tag) > 0 {
				// check if the tag equals to the data
				if r.Data == s.Tag {
					// checks if the tag is accompanied with attributes/ selectors
					if len(s.Attr) > 0 || len(s.Selector) > 0 {
						search(r)

					} else { // appends all named tag that has been chosen
						nodes = append(nodes, r)

					}
				}
			} else { // used if only attr are present for the search
				search(r)

			}
		}

		// doing a depth first search
		for c := r.FirstChild; c != nil; c = c.NextSibling {
			f(c, s)

		}
	}

	f(r.Node, s)
	l.Nodes = nodes
}

func compareQueryAttr(n html.Attribute, key string, val string) bool {
	if key == n.Key {
		if len(val) > 0 {
			if ContainsAttrString(n.Val, val) {
				return true
			}
		} else {
			return true
		}
	}

	return false
}

func compareQueryAttrList(r *html.Node, a []Attribute) bool {
	for _, x := range a {
		for _, y := range r.Attr {
			if compareQueryAttr(y, x.Name, x.Value) {
				return true
			}

		}
	}

	return false
}

func compareQuerySearch(attr html.Attribute, s Search, r *html.Node) bool {

	var attrSearch = func(as []Attribute) bool {
		for _, val := range as {
			if compareQueryAttr(attr, val.Name, val.Value) {
				return true
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

func strictQueryAttrCompare(r *html.Node, s Search) bool {
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

	return count == 0
}

func strictQuerySearch(r *html.Node, s Search) bool {
	return strictQueryCompare(r.Attr, s)
}

func queryStrictly(r *html.Node, s Search, l *NodeList, once bool) {
	isTag := len(s.Tag) > 0
	isSelector := len(s.Selector) > 0

	var f func(r *html.Node, s Search) bool
	nodes := make([]*html.Node, 0, 10)

	f = func(r *html.Node, s Search) bool {
		// check if it's an element node
		if r.Type == html.ElementNode {
			// check if it equals the tag
			if isTag {
				// current node is equal to our tag?
				if r.Data == s.Tag {
					if strictQuerySearch(r, s) {
						nodes = append(nodes, r)
						if once {
							return true
						}
					}
				}
			} else if isSelector {
				// runs when only selector present
				if strictQuerySearch(r, s) {
					nodes = append(nodes, r)
					if once {
						return true
					}
				}
			} else {
				if strictQueryAttrCompare(r, s) {
					nodes = append(nodes, r)
					if once {
						return true
					}
				}
			}
		}

		for c := r.FirstChild; c != nil; c = c.NextSibling {
			b := f(c, s)
			if b {
				return true
			}
		}
		return false
	}

	f(r, s)

	l.Nodes = nodes
}

// func queryStrictly(r *html.Node, s Search, l *NodeList, once bool) {
// 	isTag := len(s.Tag) > 0
// 	isSelector := len(s.Selector) > 0

// 	var f func(r *html.Node, s Search)
// 	nodes := make([]*html.Node, 0, 10)

// 	f = func(r *html.Node, s Search)  {
// 		// check if it's an element node
// 		if r.Type == html.ElementNode {
// 			// check if it equals the tag
// 			if isTag {
// 				// current node is equal to our tag?
// 				if r.Data == s.Tag {
// 					if strictQuerySearch(r, s) {
// 						nodes = append(nodes, r)
// 					}
// 				}
// 			} else if isSelector {
// 				// runs when only selector present
// 				if strictQuerySearch(r, s) {
// 					nodes = append(nodes, r)
// 				}
// 			} else {
// 				if strictQueryAttrCompare(r, s) {
// 					nodes = append(nodes, r)
// 				}
// 			}
// 		}

// 		for c := r.FirstChild; c != nil; c = c.NextSibling {
// 			f(c, s)

// 		}
// 	}

// 	f(r, s)

// 	l.Nodes = nodes
// }

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

func find(r *Node, s Search, l *NodeList, once bool) {
	var f func(r *html.Node, s Search) bool
	nodes := make([]*html.Node, 0)

	f = func(r *html.Node, s Search) bool {
		if r.Type == html.ElementNode {

			if len(s.Tag) > 0 {
				if s.Tag == r.Data {
					if findAttr(r.Attr, s) {
						nodes = append(nodes, r)
						if once {
							return true
						}
					}

				}
			} else if len(s.Attr) > 0 {
				if findAttr(r.Attr, s) {
					nodes = append(nodes, r)
					if once {
						return true
					}
				}
			}
		}

		for c := r.FirstChild; c != nil; c = c.NextSibling {
			b := f(c, s)
			if b {
				return true
			}
		}
		return false
	}

	f(r.Node, s)

	l.Nodes = nodes
}

// func find(r *Node, s Search, l *NodeList, once bool) {
// 	var f func(r *html.Node, s Search)
// 	nodes := make([]*html.Node, 0)

// 	f = func(r *html.Node, s Search)  {
// 		if r.Type == html.ElementNode {

// 			if len(s.Tag) > 0 {
// 				if s.Tag == r.Data {
// 					if findAttr(r.Attr, s) {
// 						nodes = append(nodes, r)

// 					}

// 				}
// 			} else if len(s.Attr) > 0 {
// 				if findAttr(r.Attr, s) {
// 					nodes = append(nodes, r)

// 				}
// 			}
// 		}

// 		for c := r.FirstChild; c != nil; c = c.NextSibling {
// 			f(c, s)

// 		}

// 	}

// 	f(r.Node, s)

// 	l.Nodes = nodes
// }

// ----------------------------------------------
func attrvalCheck(r html.Attribute, s Search) bool {
	for _, y := range s.Attr {
		if y.Name == r.Key {
			if y.Value == r.Val {
				return true
			}
		}
	}
	return false
}

func checkTag(tag, nodeTag string) bool {
	// if there is not tag or there is a tag match --> return true
	return tag == "" || tag == nodeTag
}

func findStrictly(r *html.Node, s Search, l *NodeList, once bool) {
	var nodes = make([]*html.Node, 0)
	var f func(r *html.Node, s Search) bool
	f = func(r *html.Node, s Search) bool {
		if r.Type == html.ElementNode && checkTag(s.Tag, r.Data) {
			matchedAttr := 0
			for _, x := range r.Attr {
				if attrvalCheck(x, s) {
					matchedAttr++
				}
			}
			if matchedAttr == len(r.Attr) && matchedAttr == len(s.Attr) && matchedAttr > 0 {
				nodes = append(nodes, r)
				if once {
					return true
				}
			}
			matchedAttr = 0
		}

		for c := r.FirstChild; c != nil; c = c.NextSibling {
			b := f(c, s)
			if b {
				return true
			}
		}
		return false
	}

	f(r, s)
	l.Nodes = nodes

}

// func findStrictly(r *html.Node, s Search, l *NodeList, once bool) {
// 	var nodes = make([]*html.Node, 0)
// 	var f func(r *html.Node, s Search)
// 	f = func(r *html.Node, s Search) {
// 		if r.Type == html.ElementNode && checkTag(s.Tag, r.Data) {
// 			matchedAttr := 0
// 			for _, x := range r.Attr {
// 				if attrvalCheck(x, s) {
// 					matchedAttr++
// 				}
// 			}
// 			if matchedAttr == len(r.Attr) && matchedAttr == len(s.Attr) && matchedAttr > 0 {
// 				nodes = append(nodes, r)
// 			}
// 			matchedAttr = 0
// 		}

// 		for c := r.FirstChild; c != nil; c = c.NextSibling {
// 			f(c, s)

// 		}

// 	}

// 	f(r, s)
// 	l.Nodes = nodes

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
