package req

import (
	"context"
	"net/http"
	"sync"

	"golang.org/x/net/html"
)

// Store stores the results that have been successful
type Store struct {
	// mu is a Mutex used to store multiple scraped information from multiple requests
	//
	// All successful requests and handled HTML's are stored here
	// Counter tracks the number of successful results
	mu sync.Mutex

	// res stores all information in the form of structs
	//
	// Any data that is saved should be added to a self-made struct
	// It is up to the developer to type cast the structs in order to retrieve the information
	res []interface{}

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

// Jar holds lists of clients (proxies) and headers (User-Agents)
type Jar struct {
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

// Send is a structure that determines how requests should work
//
// Holds scraping method
type Send struct {
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
	//
	// if the scrape functionality is used in Result (Result.(Scrape) / Result.(ScrapeStruct))
	// It should for dependent scraping. This means that the success of the request/scrape should incur another
	// request when needed
	Scrape func(rp Result) bool
}

// Result is used in conjunction with the method scrape in Send
//
// You can scrape, store scraped data and request new links
type Result struct {
	// document is the struct that allows scraping of webpages
	document *html.Node

	// save saves the scraped data/any data in the form of structs
	save *Store

	// ms is Send list with mutex functionality
	ms *MutexSend
}

// MutexSend stores new Sends that are made through the method Scrape in Send
//
// We need to store RS (Send) in order for us to cancel new Request objects when needed
type MutexSend struct {
	mu   sync.Mutex
	list []*Send

	// manual is a parameter that gives control over to the user on how to handle new requests
	// this should be left to default (false) if you don't want to handle storing Send structs
	manual bool

	// end notifies that no more scraping should take place
	end bool

	// passes Send to start scraping b
	scrape chan *Send
}

// Collection holds all the necessary data to request and scrape webpages
//
// This data structure allows it so to be used in a pool of other collections and can be accessed by third party
// handlers in order to control the scraping
//
// Collections should not be dereferenced when possible. There is no mutex lock functionality. This is up to the
// developer to be careful on how to access the collection
//
// It's best to use the in-house functions to form the collections. To create a safe and fully functiontion collection,
// reading documentatio on how requests, rotation and scraping is conducted in this package should be read first
type Collection struct {

	/* ---------- POOL Usage -------------- */
	// Interaction with the pool allows safe cancellations and retrieval of collections when needed

	// Identity is used for collection naming and allows the pool to identify
	Identity string

	// Cancel struct ends REQUESTS for this collection
	Cancel chan struct{}

	// Notify tells the pool that this collection REQUESTS has finally stopped running.
	// Sends the identity of this collection back to the pool
	Notify chan string

	// Complete tells the pool that this collection has finished
	Complete chan string

	// Start is used for cancelling collections. We need to know if the process has started or it will hang when trying to cancel
	Start bool

	// Done states if collection if it is has stopped processing
	// This tells the pool that it should not attempt to cancel the collection when it already has been cancelled/finished
	Done bool

	/* ---------- METHOD usage ------------ */
	// All the information needed to send requests as well as all client information linked to the collection

	// RJ stores headers and clients
	RJ *Jar

	// RS stores grouped clients, requests, retries, scraping method
	RS []*Send

	// Result contains all scraped results in this collection
	Result *Store

	// muxrs handles new RS items being created and stores them in muxrs
	muxrs *MutexSend
}

// CompleteHandler is the handler for the complete request function
//
// It defines how requests should be handled once they are finished or an error has occured
type SessionHandler interface {
	Handle(item *Send, wg *sync.WaitGroup, retry chan *Send, rc *Collection)
}

// BatchHandler is the handler for the batch request function
//
// It defines how requests should be handled once they are finished or an error has occured
type BatchHandler interface {
	Handle(item *Send, wg *sync.WaitGroup, q *Queue, rc *Collection)
	Done() bool
}

// Default handler for Complete
type DefaultSessionHandler struct {
	c int
	h int
}

// Default handler for Batch
type DefaultBatchHandler struct {
	c    int
	h    int
	done int
}

/* ?Collection ----------------------------------------------------- */

// SetMuxSend instantiates MutexSend if there is no mutexsend already there
func (rc *Collection) SetMuxSend() {
	if rc.muxrs == nil {
		rc.muxrs = &MutexSend{
			mu:   sync.Mutex{},
			end:  false,
			list: make([]*Send, 0),
		}
	}
}

// GetMuxSend returns MutexSend
func (rc *Collection) GetMuxSend() *MutexSend {
	return rc.muxrs
}

/* ?MutexSend -----------------------------------------------------  */

// New instantiates MutexSend
func (ms *MutexSend) New(scr chan *Send) *MutexSend {
	return &MutexSend{
		mu:     sync.Mutex{},
		end:    false,
		list:   make([]*Send, 0),
		scrape: scr,
	}
}

// SetChannel is used where we want the mutexsend to start a request in a particular routine
//
// The retry/try channel set by the request methods create their own retry/try channel which muxsend
// will use to carry out request conditions
func (ms *MutexSend) SetChannel(scr chan *Send) {
	ms.scrape = scr
}

// Add will add as many Send requests to the list
//
// If it has been called to end all requests, all new requests will not be carried out but will be saved in the list
func (ms *MutexSend) Add(items ...*Send) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if !ms.manual {
		ms.list = append(ms.list, items...)
	}
	if !ms.end {
		for _, x := range items {
			ms.scrape <- x
		}
	}
}

// End cancels all requests
//
// Sets end to true stating that no more scraping should continue
func (ms *MutexSend) End() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.end = true
	for _, x := range ms.list {
		c := *x.Request.Cancel
		c()
	}
}

// Manual gives control to the developer on how Send structs should be stored
func (ms *MutexSend) Manual() {
	ms.manual = true
}

/* ?Send ----------------------------------------------------- */

// Decrement the amount of retries
func (rs *Send) Decrement() {
	rs.Retries--
}

// AddHeader adds any header key-value pairs to Send
func (rs *Send) AddHeader(key, value string) {
	rs.Request.Request.Header.Add(key, value)
}

// SetHost sets Host for request headers
func (rs *Send) SetHost(host string) {
	rs.Request.Request.Host = host
}

// SetHeaders directly attaches parameter headers to Send struct
func (rs *Send) SetHeaders(headers map[string][]string) {
	rs.Request.Request.Header = headers
}

// SetHeadersStruct applies header to the Send struct
func (rs *Send) SetHeadersStruct(header *http.Header) {
	rs.Request.Request.Header = *header
}

/* ?Store ----------------------------------------------------- */

// Add stores struct into a list - saving for later use
func (c *Store) Add(b interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.res = append(c.res, b)
	c.counter++
}

// Read returns interface list full of structs
func (c *Store) Read() []interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.res
}

// Count return number of added data
func (c *Store) Count() int {
	return c.counter
}

/* ?Result ----------------------------------------------------- */

// New instantiates Result struct with scraping document, saving struct, and mutexed sending requests
func (rp Result) New(doc *html.Node, save *Store, mutexSend *MutexSend) Result {
	rp.document = doc
	rp.save = save
	rp.ms = mutexSend
	return rp
}

// Document returns scraped document
func (rp Result) Document() *html.Node {
	return rp.document
}

// Save to list
func (rp Result) Save(item interface{}) {
	rp.save.Add(item)
}

// Scrape allows continued requests depending on scraped results
func (rp Result) Scrape(m func(rp Result) bool, url ...string) {
	rp.ms.Add(ItemToSend(MakeRequestItems(ConvertToURL(url)), m)...)
}

// ScrapeStruct allows you to send structs instead of just a url
func (rp Result) ScrapeStruct(rs ...*Send) {
	rp.ms.Add(rs...)
}

/* ?RequestItem ----------------------------------------------------- */

// CancelRequest cancels the current request with context
func (ri *RequestItem) CancelRequest() {
	cancel := *ri.Cancel
	cancel()
}

/* HANDLERS ----------------------------------------------------- */

// Handle requests
func (dh *DefaultSessionHandler) Handle(item *Send, wg *sync.WaitGroup, retry chan *Send, rc *Collection) {
	switch {
	case item.Caught:
		wg.Add(1)
		dh.c = changeHeaders(item.Request.Request, rc.RJ, dh.c)
		item.Caught = false
		go HandleRequest(true, item, retry, rc.Result, rc.muxrs, wg)
		break
	case item.Retries == 0:
		break
	default:
		wg.Add(1)
		dh.h = changeClient(item, rc.RJ.Clients, dh.h)
		go HandleRequest(true, item, retry, rc.Result, rc.muxrs, wg)
		break
	}
}

// Handle requests
func (dh *DefaultBatchHandler) Handle(item *Send, wg *sync.WaitGroup, q *Queue, rc *Collection) {
	if item.Caught {
		dh.h = changeHeaders(item.Request.Request, rc.RJ, dh.h)
		item.Caught = false
		q.Add(item)
	} else if item.Retries == 0 {
		dh.done--
	} else {
		dh.c = changeClient(item, rc.RJ.Clients, dh.c)
		q.Add(item)
	}
}

// Done returns bool if all requests have been done
func (dh *DefaultBatchHandler) Done() bool {
	return dh.done == 0
}

/* METHODS ----------------------------------------------------- */

// Scrape converts a string of urls into Send struct
func Scrape(url []string) []*Send {
	urls := MakeRequestItems(ConvertToURL(url))
	items := make([]*Send, 0, len(urls))
	for _, x := range urls {
		items = append(items, &Send{
			Request: x,
			Retries: 1,
		})
	}
	return items
}
