package req

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// Proxy constants
const (
	ProxyConnectionPoolError = "ProxyConnection: Client Failed! [POOL]"
	sleepTime                = 1000
)

// ----------------------------------------------------------

// CompleteSession handles retries and changing headers and proxies
// Handlers can be used to handle unsuccessful requests
func Session(rc *Collection, handle SessionHandler) {

	var wg sync.WaitGroup
	var handler func(item *Send, wg *sync.WaitGroup, retry chan *Send, rc *Collection)
	var dh *DefaultSessionHandler

	retry := make(chan *Send, 10) // Requests for those that need a retry or they have finsihed retrying
	end := make(chan struct{})

	collectionChecker(rc)
	result := rc.Result     // Scrape results
	cancel := rc.Cancel     // Pool usage
	complete := rc.Complete // Pool usage
	ms := rc.muxrs
	ms.SetChannel(retry)

	if handle != nil {
		handler = handle.Handle
	} else {
		dh = &DefaultSessionHandler{}
		handler = dh.Handle
	}

	for _, x := range rc.RS {
		wg.Add(1)
		go HandleRequest(true, x, retry, result, ms, &wg)
	}

	go func() {
		wg.Wait()
		if complete != nil {
			complete <- rc.Identity
		}
		close(end)
		rc.Done = true
		return
	}()

loop:
	for {
		select {
		case item := <-retry:
			handler(item, &wg, retry, rc)
		case <-cancel:
			break loop
		case <-end:
			break loop
		}
	}

	return
}

// ----------------------------------------------------------

// Simple handles requests normally. It can take in proxied clients or not and doesn't
// allow retries.
func Run(rc *Collection) {
	var wg sync.WaitGroup

	// Retry functionality is set to continue scraping new calls and NOT failed calls
	// However its can be setup in any way
	retry := make(chan *Send)
	finish := make(chan struct{})

	collectionChecker(rc)
	result := rc.Result
	cancel := rc.Cancel
	complete := rc.Complete
	ms := rc.muxrs
	ms.SetChannel(retry)

	for _, x := range rc.RS {
		wg.Add(1)
		go HandleRequest(false, x, retry, result, ms, &wg)
	}

	go func() {
		wg.Wait()
		close(finish)
		if complete != nil {
			complete <- rc.Identity
		}
		// CONFLICT
		rc.Done = true
		return
	}()

loop:
	for {
		select {
		case <-cancel:
			break loop
		case <-finish:
			break loop
		case item := <-retry:
			wg.Add(1)
			go HandleRequest(false, item, retry, result, ms, &wg)
			break
		}
	}

	return
}

// ----------------------------------------------------------

// HandleRequest carries out request and also runs scrape
func HandleRequest(enforce bool, req *Send, retry chan *Send, rr *Store, ms *MutexSend, wg *sync.WaitGroup) {

	var resp *http.Response
	var err error

	if wg != nil {
		defer wg.Done()
	}

	client := req.Client
	request := req.Request.Request

	if enforce {
		resp, err = enforceTrue(client, request, req, retry)
	} else {
		resp, err = enforceFalse(client, request)
	}

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	data, err := RunScrape(resp, rr, ms, req.Scrape)
	if err != nil {
		log.Println("Couldn't read body")
		return
	}

	// if we are not enforcing, we can leave right now
	if !enforce {
		return
	}

	// if the request has gone successfull but the scrape has not i.e. website has detected it's being scraped and providing non-essential information
	if !data {
		req.Caught = true
		retry <- req
		return
	}

	req.Retries = 0
	retry <- req

	return
}

// enforce makes it so that retries/header change/ip change are enforced
func enforceTrue(client *http.Client, request *http.Request, req *Send, retry chan *Send) (*http.Response, error) {
	resp, err := client.Do(request)

	if err != nil {
		req.Decrement()
		retry <- req
		return nil, errors.New(ProxyConnectionPoolError)
	}

	return resp, nil
}

// notEnforce makes it so that only one single request is made per item/ no proxies/clients are not needed
func enforceFalse(client *http.Client, request *http.Request) (*http.Response, error) {
	var cli http.Client

	if client != nil {
		cli = *client
	} else {
		cli = http.Client{}
	}

	resp, err := cli.Do(request)

	if err != nil {
		return nil, errors.New(ProxyConnectionPoolError)
	}

	return resp, nil
}

// ----------------------------------------------------------

// changeClient changes client if condition requires to do so
//
// NOTE - This default implementation should be changed using Handlers
func changeClient(client *Send, list []*http.Client, counter int) int {

	newCli := list[counter]

	client.Client = newCli
	counter++

	if counter == len(list) {
		counter = 0
	}

	return counter
}

// changeHeaders changes headers if conditions requires to do so
//
// NOTE - This default implementation should be changed using Handlers
func changeHeaders(req *http.Request, jar *Jar, count int) int {
	req.Header = *jar.Headers[count]
	count++
	if count == len(jar.Headers) {
		return 0
	}
	return count
}

// RunScrape runs the scraping method/document
// if false, it means that the scrape was unsuccessful
// if true, scrape successful
func RunScrape(r *http.Response, res *Store, ms *MutexSend, m func(rp Result) bool) (bool, error) {
	defer r.Body.Close()
	utf8set, err := charset.NewReader(r.Body, r.Header.Get("Content-Type"))
	if err != nil {
		log.Println("Failed utf8set")
		return false, err
	}
	bytes, err := ioutil.ReadAll(utf8set)
	if err != nil {
		log.Println("Failed ioutil")
		return false, err
	}

	// item := traverse.HTMLDocBytes(&bytes)
	// pack := Result{}
	// pack = pack.New(item, res, ms)
	pack := Result{}
	parse, _ := htmlParser(bytes)
	pack = pack.New(parse, res, ms)
	if m == nil {
		return true, err
	}
	return m(pack), err
}

// collectionChecker fills in any missing structs needed to run the collection
//
// NOTE - This is primarily used to make sure that all request methods e.g. CompleteSession, Simple
// can be run without using a Pool
// If the correct components are not added to the collection, the intended scrape might not go as planned
func collectionChecker(rc *Collection) *Collection {
	if rc.Cancel == nil {
		rc.Cancel = make(chan struct{}, 1)
	}

	if rc.RJ == nil {
		rc.RJ = &Jar{}
	}

	if rc.RS == nil {
		rc.RS = make([]*Send, 0)
	}

	if rc.Result == nil {
		rc.Result = &Store{}
	}

	if rc.muxrs == nil {
		rc.muxrs = &MutexSend{}
	}

	return rc
}

func htmlParser(b []byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(b))
}
