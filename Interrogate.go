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

var (
	f func(r *html.Node, s Search) bool
)

// ----------------------------------------------------------------

func query(r *html.Node, s Search) []*html.Node {
	var f func(r *html.Node, s Search)
	nodes := make([]*html.Node, 0, 5)
	var searchWords [][]string

	for _, x := range s.Attr {
		word := BreakWords(x.Value)
		searchWords = append(searchWords, word)
	}

	f = func(r *html.Node, s Search) {
		if r.Type == html.ElementNode {
			if s.Tag != "" {
				if r.Data == s.Tag {
					if querySearch(r.Attr, s, searchWords) {
						nodes = append(nodes, r)
					}
				}
			} else {
				if querySearch(r.Attr, s, searchWords) {
					nodes = append(nodes, r)
				}
			}
		}

		for c := r.FirstChild; c != nil; c = c.NextSibling {
			f(c, s)
		}

	}

	f(r, s)
	return nodes
}

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

func find(r *html.Node, s Search, once bool) []*html.Node {
	// var f func(r *html.Node, s Search) bool
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

	f(r, s)

	return nodes
}

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

func findStrictly(r *html.Node, s Search, once bool) []*html.Node {
	var nodes = make([]*html.Node, 0)
	selectorAttrlen := len(s.Attr) + len(s.Selector)
	// var f func(r *html.Node, s Search) bool
	f = func(r *html.Node, s Search) bool {
		if r.Type == html.ElementNode && checkTag(s.Tag, r.Data) {
			if len(r.Attr) == selectorAttrlen {
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

	return nodes
}

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

func BreakWords(str string) []string {
	var list []string
	list = append(list, strings.Fields(str)...)
	return list
}

// --------------------------------------------------------------------------------------------
//  HTML parsing. Independeant from the DocumentGroup

func Parse(r io.Reader) (*html.Node, error) {
	return html.Parse(r)
}

func EasyParse(s *[]byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(*s))
}
