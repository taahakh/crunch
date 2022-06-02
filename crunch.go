package crunch

import (
	"io"
	"net/http"

	"github.com/taahakh/crunch/traverse"
)

// Docify converts any file into nodes that can be traversed
func Docify(r io.Reader) (*traverse.HTMLDocument, error) {
	doc, err := traverse.HTMLDoc(r)
	return doc, err
}

// Takes any http response and docify's it
func ResponseToDoc(r *http.Response) (traverse.HTMLDocument, error) {
	doc, err := traverse.HTMLDocUTF8(r)
	return doc, err
}

// Run traversal on document
// Available search strings functions
// --> query, queryOnce, find, findOnce, findStrict, findStrictOnce
func Find(function, search string, doc *traverse.HTMLDocument) *traverse.HTMLDocument {
	traverse.SearchSwitch(function, search, doc)
	return doc
}

// Run custom code on what to do when you have found the particular information
func Manipulate(doc *traverse.HTMLDocument, method func(doc *traverse.HTMLDocument)) {
	method(doc)
}

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
