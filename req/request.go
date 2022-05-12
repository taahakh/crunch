package req

import (
	"context"
	"net/http"
	"sync"

	"github.com/taahakh/speed/traverse"
)

// Stores the results that have been successful
type RequestResult struct {
	// All successful requests and handled HTML's are stored here
	// Counter tracks the number of successful results
	mu      sync.Mutex
	res     []*interface{}
	counter int
}

type RequestItem struct {
	Request *http.Request
	Cancel  *context.CancelFunc
}

// Groups of ips and links
type RequestJar struct {
	Clients []*http.Client
	Links   []*RequestItem

	// For mainly user-agents and also headers
	Headers []*http.Header
}

type RequestSend struct {
	// Retries <= 0  - tries unless finished
	Client  *http.Client
	Request *RequestItem

	// Change user-agents if its has been detected non-human
	Caught bool

	// Failed ip request - how many more rotations to do
	Retries int

	// The method to use to scrape the webpage
	// Method func(doc *traverse.HTMLDocument, rr *RequestResult) bool
	Method func(rp ResultPackage) bool
}

type ResultPackage struct {
	document     *traverse.HTMLDocument
	save         *RequestResult
	scrape       func(url string) *RequestSend
	scrapeStruct func(rs *RequestSend) *RequestSend
}

type RequestCollection struct {
	// Finish tells us when we want the webscrape to end by no matter what
	// Finish nil will go on until everything is finished

	/* ---------- POOL Usage -------------- */
	// Interaction with the pool allows safe cancellations and retrieval of collections when needed

	// Provides identity for collection and allows the pool to identify
	Identity string

	// Pool sending cancel struct to end REQUESTS for this collection
	Cancel chan struct{}

	// Telling the pool that this collection REQUESTS has finally stopped running.
	// Sends the identity of this collection back to the pool
	Notify *chan string

	// Tells the pool that this collection has finished
	Complete *chan string

	// Used for cancelling collections. We need to know if the process has started or it will hang when trying to cancel
	Start bool

	// State of collection if it is has stopped processing
	// This tells the pool that it should not attempt to cancel the collection when it already has been cancelled/finished
	Done bool

	/* ---------- METHOD usage ------------ */
	// All the information needed to send requests as well as all client information linked to the collection
	// RequestJar needs to be MUTEXED

	RJ     *RequestJar
	RS     []*RequestSend
	Result *RequestResult
}

// Clients are the proxies and engine
// Links are the request we want to make
func (rj *RequestJar) CreateHandle(rt int) []*RequestSend {
	rs := make([]*RequestSend, 0, len(rj.Links))
	counter := 0
	for _, x := range rj.Links {
		rs = append(rs, &RequestSend{
			Request: x,
			Client:  rj.Clients[counter],
			// Retries: rj.Req[i].Retries,
			Retries: rt,
		})
		counter++
		if counter == len(rj.Clients) {
			counter = 0
		}
	}

	return rs
}

func (rs *RequestSend) Decrement() {
	rs.Retries--
}

func (rs *RequestSend) AddHeader(key, value string) {
	rs.Request.Request.Header.Add(key, value)
}

func (rs *RequestSend) SetHost(host string) {
	rs.Request.Request.Host = host
}

func (rs *RequestSend) SetHeaders(headers map[string][]string) {
	rs.Request.Request.Header = headers
}

func (rs *RequestSend) SetHeadersStruct(header *http.Header) {
	rs.Request.Request.Header = *header
}

func (c *RequestResult) Add(b interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.res = append(c.res, &b)
	c.counter++
}

func (c *RequestResult) Read() []*interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.res
}

func (c *RequestResult) Count() int {
	return c.counter
}

// func (r *RequestCollection) SignalFinish() {
// 	fmt.Println("Signaled finish: ", r.Identity)
// 	r.Done = true
// 	*r.Notify <- r.Identity
// }

func (ri *RequestItem) CancelRequest() {
	cancel := *ri.Cancel
	cancel()
}

func (rp ResultPackage) New(doc *traverse.HTMLDocument, save *RequestResult, scrape func(url string) *RequestSend, scrapeStruct func(rs *RequestSend) *RequestSend) ResultPackage {
	rp.document = doc
	rp.save = save
	rp.scrape = scrape
	rp.scrapeStruct = scrapeStruct
	return rp
}

func (rp ResultPackage) Document() *traverse.HTMLDocument {
	return rp.document
}

func (rp ResultPackage) Save(item interface{}) {
	rp.save.Add(item)
}

func (rp ResultPackage) Scrape(url string) {
	rp.scrape(url)
}

func (rp ResultPackage) ScrapeStruct(rs *RequestSend) {
	rp.scrapeStruct(rs)
}
