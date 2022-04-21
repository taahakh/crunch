package req

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/taahakh/speed/traverse"
)

// type RJ interface {
// 	Clients() []*http.Client
// 	Requests() []*http.Request
// }

// type RS interface {
// 	Client() *http.Client
// 	Request() *http.Request
// 	Cancel() *context.CancelFunc
// 	Retries() int
// }

// type Result interface {
// 	Mutex() sync.Mutex
// 	Counter() int

// }

// Stores the results that have been successful
type RequestResult struct {
	// All successful requests and handled HTML's are stored here
	// Counter tracks the number of successful results
	mu      sync.Mutex
	res     []traverse.HTMLDocument
	counter int
}

type RequestItem struct {
	Request *http.Request
	Cancel  *context.CancelFunc
}

// Groups of ips and links
type RequestJar struct {
	Clients []*http.Client
	Links   []*RequestItem // this is intially in the form of url.URL but is then converted to string
}

type RequestSend struct {
	// Retries <= 0  - tries unless finished
	Client  *http.Client
	Request *RequestItem
	Retries int
}

type RequestCollection struct {
	// Finish tells us when we want the webscrape to end by no matter what
	// Finish nil will go on until everything is finished

	/* ---------- POOL Usage -------------- */
	// Interaction with the pool allows safe cancellations and retrieval of collections when needed

	// Provides identity for collection and allows the pool to identify
	Identity string

	// Pool sending cancel struct to end goroutines/requests for this collection
	Cancel chan struct{}

	// Telling the pool that this collection has finally stopped running.
	// Sends the identity of this collection back to the pool
	Notify *chan string

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

func (c *RequestResult) add(b traverse.HTMLDocument) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.res = append(c.res, b)
	c.counter++
}

func (c *RequestResult) Read() []traverse.HTMLDocument {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.res
}

func (c *RequestResult) Count() int {
	return c.counter
}

func (r *RequestCollection) SignalFinish() {
	fmt.Println("Signaled finish: ", r.Identity)
	*r.Notify <- r.Identity
	// *r.Notify <- r

}

func (ri *RequestItem) CancelRequest() {
	cancel := *ri.Cancel
	cancel()
}
