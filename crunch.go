package crunch

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/taahakh/crunch/req"
	"github.com/taahakh/crunch/traverse"
)

// Docify converts any file into nodes that can be traversed
// A wrapper to the function traverse.HTMLDoc
func Docify(r io.Reader) (*traverse.HTMLDocument, error) {
	doc, err := traverse.HTMLDoc(r)
	return doc, err
}

// Takes any http response and docify's it
// A wrapper to the function traverse.HTMLDocUTF8
func ResponseToDoc(r *http.Response) (traverse.HTMLDocument, error) {
	doc, err := traverse.HTMLDocUTF8(r)
	return doc, err
}

// Run traversal on document
// A wrapper to the function traverse.SearchSwitch
// Available search strings functions
// --> query, queryOnce, find, findOnce, findStrict, findStrictOnce
func Find(function, search string, doc *traverse.HTMLDocument) *traverse.HTMLDocument {
	traverse.SearchSwitch(function, search, doc)
	return doc
}

/* CONNECTION */

// Proxy is a simple request made with a proxy
//
// ONLY HTTP/HTTPS compatable
func Proxy(link, proxy string, timeout time.Duration) (*http.Response, error) {
	p, err := url.Parse(proxy)
	if err != nil {
		log.Println("Proxy parsing not working")
		return &http.Response{}, err
	}

	l, err := url.Parse(link)
	if err != nil {
		log.Println("Link parsing not working")
		return &http.Response{}, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(p),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	req, err := http.NewRequest("GET", l.String(), nil)
	if err != nil {
		log.Println("New request failed")
		return &http.Response{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println("Client do not working")
		return &http.Response{}, err
	}

	return res, err
}

// NoProxy doesn't use proxy to make requests
//
// Uses shared client for all requests
func NoProxy(urls []string, timeout time.Duration, method func(rp req.Result) bool) *req.Collection {
	rs := make([]*req.Send, 0, len(urls))
	reqe := req.ConvertToURL(urls)

	client := &http.Client{}

	ri := req.MakeRequestItems(reqe)
	for _, x := range ri {

		rs = append(rs, &req.Send{
			Request: x,
			Scrape:  method,
			Client:  client,
		})
	}

	return &req.Collection{
		RS: rs,
		RJ: &req.Jar{
			Clients: []*http.Client{{}},
		},
	}
}

// ProxySetup allows requests to be dialled through a proxy
//
// Proxy, Url String lists and timeout must not be empty
// Method func and retries should be added to make the request and scrape more useful
func ProxySetup(
	proxy []string,
	urls []string,
	headers []*http.Header,
	retries int,
	timeout time.Duration,
	method func(rp req.Result) bool) *req.Collection {

	var ri []*req.RequestItem

	if proxy == nil || urls == nil || len(proxy) == 0 || len(urls) == 0 {
		// return nil
		panic("Proxy or Urls must not be nil")
	}

	reqe := req.ConvertToURL(urls)
	cli := req.ConvertToURL(proxy)

	ri = req.MakeRequestItems(reqe)
	c := req.MakeProxyClients(cli, timeout)

	if headers != nil {
		reqX, err := req.ApplyHeadersRI(ri, headers)
		if err != nil {
			log.Println("Headers could not be applied")
		}
		ri = reqX
	}

	rj := &req.Jar{
		Clients: c,
		Headers: headers,
	}

	rs, err := req.MakeSends(c, ri, retries, method)
	if err != nil {
		panic("NICE")
	}

	return &req.Collection{
		RJ: rj,
		RS: rs,
	}
}

// Run the scrape using a select request option
// --> Session [Requires SessionHandler], Run
func Do(col *req.Collection, request string, handler req.SessionHandler) (*req.Collection, error) {
	switch request {
	case "Run":
		req.Run(col)
		break
	case "Session":
		req.Session(col, handler)
		break
	default:
		return nil, errors.New("Request String was not found")
	}

	return col, nil
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
