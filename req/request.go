package req

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/taahakh/speed/traverse"
)

// Groups of ips and links
type RequestJar struct {
	Clients []*http.Client
	Links   []*http.Request // this is intially in the form of url.URL but is then converted to string
}

type RequestSend struct {
	// Retries <= 0  - tries unless finished
	Client  *http.Client
	Request *http.Request
	Cancel  *context.CancelFunc
	Retries int
}

type RequestCollection struct {
	// Finish tells us when we want the webscrape to end by no matter what
	// Finish nil will go on until everything is finished

	/* ---------- POOL Usage -------------- */
	// Interaction with the pool allows safe cancellations and retrieval of collections when needed

	Identity string        // Provides identity
	Safe     chan struct{} // Telling the pool when it is safe to exit cancel for further use. Done is set to true. DEPRECIATED SOON?
	Notify   *chan string  // Telling the pool that this collection has stopped running. Sends the identity of this collection back to the pool

	/* ---------- METHOD usage ------------ */
	// All the information needed to send requests as well as all client information linked to the collection
	// RequestJar needs to be MUTEXED

	RJ     *RequestJar
	RS     []*RequestSend
	Result *RequestResult
	Cancel chan struct{} // cancel channel to end goroutines/requests for this collection
	// Extend chan *RequestSend // Used for retries and extending the collection
	Finish string // how long it should take before the rc should end. Should follow time.Duration rules to get desired result
	Done   bool   // State when this is done. This is also POOL usage. DEPRECIATED SOON?
}

// Stores the results that have been successful
type RequestResult struct {
	// All successful requests and handled HTML's are stored here
	// Counter tracks the number of successful results
	mu      sync.Mutex
	res     []traverse.HTMLDocument
	counter int
}

type GroupRequest struct {
	Ips     []string
	Links   []string
	Timeout time.Duration
}

type SingleRequest struct {
	proxyClient *http.Client
	link        *http.Request
	retries     int
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

func (rt *SingleRequest) decrement() {
	rt.retries -= 1
}

func (r *RequestCollection) SignalFinish() {
	fmt.Println(r.Identity)
	*r.Notify <- r.Identity
}
