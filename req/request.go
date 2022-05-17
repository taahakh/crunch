package req

import (
	"context"
	"net/http"
	"sync"

	"github.com/taahakh/speed/traverse"
)

// RequestResult stores the results that have been successful
type RequestResult struct {
	// mu is a Mutex used to store multiple scraped information from multiple requests
	//
	// All successful requests and handled HTML's are stored here
	// Counter tracks the number of successful results
	mu sync.Mutex

	// res stores all information in the form of structs
	//
	// Any data that is saved should be added to a self-made struct
	// It is up to the developer to type cast the structs in order to retrieve the information
	res []*interface{}

	// counter tracks the number of results
	//
	// DEPRECIATED
	counter int
}

// RequestItem stores the Request struct and its cancellation function
//
// Request is attached to context.WithContext in order for it to be cancelled
// This usage is attached to the Pool struct which cancels/ends all requests
type RequestItem struct {
	Request *http.Request
	Cancel  *context.CancelFunc
}

// RequestJar holds lists of clients (proxies) and headers (User-Agents)
type RequestJar struct {
	// Clients structs holding transport information for proxies
	// Instead of clients, it should be replaced with http.Transport
	// Transports hold how the request should be made. Client is the wrapper
	// that runs the request.
	//
	// NEEDS TO BE CHANGED
	Clients []*http.Client

	// For mainly user-agents and header information
	// Needs to be in the form of []http.Header.
	// We want to copy the header map, NOT the address
	//
	// NEEDS TO BE CHANGED
	Headers []*http.Header
}

// RequestSend is a structure that determines how requests should work
//
// Holds scraping method
type RequestSend struct {
	// Client proxy
	//
	// This MUST not be EMPTY
	// This could/should be replaced with http.Transport
	Client *http.Client

	// Request struct
	Request *RequestItem

	// Caught is used to say if the User-Agents need to be changed
	//
	// Checked by the retry handler to see if the request has been caught being or bot/failed/or otherwise
	// Caught is determined by the programmer through the scrape process
	Caught bool

	// Number of tries it has attempted for a successful request
	//
	// Retries >= 1  --> Defines how many attempts
	// Retries <= -1 --> Keeps on retrying indefinietly
	// Should not be set to zero. Default set to 1
	Retries int

	// Method function set by the developer on how to scrape the website
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
	rp.ms.Add(ItemToSend(MakeRequestItems(ConvertToURL(url)), m)...)
}

func (rp ResultPackage) ScrapeStruct(rs ...*RequestSend) {
	rp.ms.Add(rs...)
}

func Scrape(url []string) []*RequestSend {
	urls := MakeRequestItems(ConvertToURL(url))
	items := make([]*RequestSend, 0, len(urls))
	for _, x := range urls {
		items = append(items, &RequestSend{
			Request: x,
			Retries: 1,
		})
	}
	return items
}
