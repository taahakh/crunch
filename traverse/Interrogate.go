package traverse

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// All searches are case sensitive. It must be exact

// For selector and Attr values
type Attribute struct {
	Name  string
	Value string
}

// type AttributeMap map[string][]string
type AttributeMap map[string]string

// A wrapper for a number of functions
// Possibility of being depreciated soon
// type NodeList struct {
// 	Nodes []*html.Node // lets users have their freedom to manipulate the raw html.Node
// }

type NodeList []*html.Node

// A wrapper to access functions
type Node struct {
	Node *html.Node
}

var (
	f func(r *html.Node, s Search) bool
)

// ----------------------------------------------------------------

// Finds words that we have given to search for. The search doesn't
// exactly match the values given but if the word exists in the string
// then it counts
func query(r *html.Node, s Search, once bool) []*html.Node {
	nodes := make([]*html.Node, 0, 5)
	var searchWords [][]string

	for _, x := range s.Attr {
		word := BreakWords(x.Value)
		searchWords = append(searchWords, word)
	}

	f = func(r *html.Node, s Search) bool {
		if r.Type == html.ElementNode {
			if s.Tag != "" {
				if r.Data == s.Tag {
					if querySearch(r.Attr, s, searchWords) {
						nodes = append(nodes, r)
						if once {
							return true
						}
					}
				}
			} else {
				if querySearch(r.Attr, s, searchWords) {
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

	fmt.Println(nodes)

	return nodes
}

// searches for all instances of the words we want to find within
// value strings attributes
func querySearch(n []html.Attribute, s Search, word [][]string) bool {
	count := len(s.Attr)

	for _, a := range n {
		for i, b := range s.Attr {
			if a.Key == b.Name {
				if len(b.Value) > 0 {
					if CompareAttrLists(word[i], a.Val) {
						count--
					}
				} else {
					count--
				}
			}
		}
	}

	return count == 0
}

// ---------------------------------------------------------------

//  Finding the correct attr-value keypair
func findAttr(r []html.Attribute, s Search) bool {

	numToBeFound := len(s.Attr)
	for _, x := range r {
		if attrvalCheck(x, s) {
			numToBeFound--
		}
	}

	return numToBeFound == 0
}

// Searches the exact string given for each value of attr
// the values given must be exact in order to find node
func find(r *html.Node, s Search, once bool) []*html.Node {
	// var f func(r *html.Node, s Search) bool
	nodes := make([]*html.Node, 0, 1)

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

	f(r, s)

	fmt.Println(nodes)

	return nodes
}

// ----------------------------------------------

func checkTag(tag, nodeTag string) bool {
	// if there is not tag or there is a tag match --> return true
	return tag == "" || tag == nodeTag
}

func findStrictly(r *html.Node, s Search, once bool) []*html.Node {
	var nodes = make([]*html.Node, 0, 1)
	var fu func(r *html.Node, s Search)
	sLen := len(s.Attr)
	selectorAttrlen := sLen + len(s.Selector)
	caught := false
	fu = func(r *html.Node, s Search) {
		if r.Type == html.ElementNode && checkTag(s.Tag, r.Data) {
			if len(r.Attr) == selectorAttrlen {
				matchedAttr := 0
				for _, x := range r.Attr {
					if attrvalCheck(x, s) {
						matchedAttr++
					}
				}
				if matchedAttr == len(r.Attr) && matchedAttr == sLen && matchedAttr > 0 {
					nodes = append(nodes, r)
					if once {
						caught = true
						return
					}
				}
				matchedAttr = 0
			}

		}

		for c := r.FirstChild; c != nil; c = c.NextSibling {
			fu(c, s)
			if caught {
				return
			}
		}
	}

	fu(r, s)

	fmt.Println(nodes)

	return nodes
}

// --------------------------------------------------------------------------------------------

// Takes already word list from search struct and now taking in
// node attr string. Comparing
func CompareAttrLists(searchList []string, node string) bool {
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

// Breaks down the sentence into a word array
// Used in query() to find words within sentences
func BreakWords(str string) []string {
	var list []string
	list = append(list, strings.Fields(str)...)
	return list
}

// Comparison check - needs attr-val keypair to be correct
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

// --------------------------------------------------------------------------------------------
//  HTML parsing. Independeant from the DocumentGroup

func EasyParse(s *[]byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(*s))
}

func StringEasyParse(s *[]byte) (*html.Node, error) {
	str := string(*s)
	return html.Parse(strings.NewReader(str))
}
