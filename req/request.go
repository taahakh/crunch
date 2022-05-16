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
	// Clients
	Clients []*http.Client

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

	/*
		The method to use to scrape the webpage
		Method func(doc *traverse.HTMLDocument, rr *RequestResult) bool
	*/
	Method func(rp ResultPackage) bool
}

type ResultPackage struct {
	document *traverse.HTMLDocument
	save     *RequestResult
	ms       *MutexSend
}

type MutexSend struct {
	mu     sync.Mutex
	list   []*RequestSend
	end    bool
	scrape chan *RequestSend
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
	// RequestSend needs to be MUTEXED

	RJ     *RequestJar
	RS     []*RequestSend
	Result *RequestResult
	muxrs  *MutexSend
}

func (rc *RequestCollection) SetMuxSend() {
	if rc.muxrs == nil {
		rc.muxrs = &MutexSend{
			mu:   sync.Mutex{},
			end:  false,
			list: make([]*RequestSend, 0),
		}
	}
}

func (rc *RequestCollection) GetMuxSend() *MutexSend {
	return rc.muxrs
}

func (ms *MutexSend) New(scr chan *RequestSend) *MutexSend {

	return &MutexSend{
		mu:     sync.Mutex{},
		end:    false,
		list:   make([]*RequestSend, 0),
		scrape: scr,
	}
}

func (ms *MutexSend) SetChannel(scr chan *RequestSend) {
	ms.scrape = scr
}

func (ms *MutexSend) Add(items ...*RequestSend) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.list = append(ms.list, items...)
	if !ms.end {
		for _, x := range items {
			ms.scrape <- x
		}
	}
}

func (ms *MutexSend) End() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.end = true
	for _, x := range ms.list {
		c := *x.Request.Cancel
		c()
	}
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

func (ri *RequestItem) CancelRequest() {
	cancel := *ri.Cancel
	cancel()
}

func (rp ResultPackage) New(doc *traverse.HTMLDocument, save *RequestResult, mutexSend *MutexSend) ResultPackage {
	rp.document = doc
	rp.save = save
	// rp.channel = retry
	rp.ms = mutexSend
	return rp
}

func (rp ResultPackage) Document() *traverse.HTMLDocument {
	return rp.document
}

func (rp ResultPackage) Save(item interface{}) {
	rp.save.Add(item)
}

func (rp ResultPackage) Scrape(m func(rp ResultPackage) bool, url ...string) {
	rp.ms.Add(ItemToSend(CreateLinkRequestContext(ConvertToURL(url)), m)...)
}

func (rp ResultPackage) ScrapeStruct(rs ...*RequestSend) {
	rp.ms.Add(rs...)
}

func Scrape(url []string) []*RequestSend {
	urls := CreateLinkRequestContext(ConvertToURL(url))
	items := make([]*RequestSend, 0, len(urls))
	for _, x := range urls {
		items = append(items, &RequestSend{
			Request: x,
			Retries: 1,
		})
	}
	return items
}
