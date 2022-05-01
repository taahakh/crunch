package req

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/taahakh/speed/traverse"
	"golang.org/x/net/html/charset"
)

const (
	sleepTime = 1000
)

// ---------------------------------------------------------------------------------

// Handles retries, individual request timeouts and cancellations,
// all request cancellations

// request - number of request to push per batch

// size - how many requests per batch
// gap - how long until to send next batch. if the time gap = 0, then we will return. there needs to be a timed batch
func Batch(rj *RequestCollection, size int, gap string) {
	defer rj.SignalFinish()

	var dur time.Duration

	// check if time duration is empty
	if gap != "" {
		t, err := time.ParseDuration(gap)
		if err != nil {
			return
		}
		dur = t
	}

	// if time duration == 0
	if dur == 0 {
		return
	}

	var wg sync.WaitGroup
	retry := make(chan *RequestSend, size)
	// Scrape results
	result := rj.Result
	// safe := rj.Safe
	// The pool telling the collection to end all requests
	cancel := rj.Cancel
	// End the goroutine where we do requests.
	// This goroutine ends when the second goroutine receives a cancellation singal from the pool
	// sends struct to end this goroutine. Cannot use the same cancellation signal to end the goroutine
	// we also want to make sure that any retries that are updated to the queue does not get missed by the cancellation
	end := make(chan struct{})

	counter := 0
	cDone := len(rj.RJ.Links)
	cHeader := 0

	q := Queue{}
	q.Make(rj.RS)

	go func() {
		// We go thorugh the list first without doing the retries
		for {
			select {
			case <-end:
				return
			default:
				for i := 0; i < size; i++ {
					item := q.Pop()
					if item == nil {
						continue
					}
					go HandleRequest(item, retry, result, &wg)
				}

				time.Sleep(dur)
			}
		}
	}()

	go func() {
		for {
			time.Sleep(time.Millisecond * sleepTime)

			select {
			case item := <-retry:
				switch {
				case item.Caught:
					cHeader = changeHeaders(item.Request.Request, rj.RJ, cHeader)
					q.Add(item)

					break
				case item.Retries == 0:
					cDone--

					break
				default:
					counter = changeClient(item.Client, rj.RJ.Clients, counter)
					q.Add(item)
				}

				break
			case <-cancel:
				end <- struct{}{}

				return
			default:
				if cDone == 0 {
					end <- struct{}{}

					return
				}
			}
		}
	}()

	return
}

func CompleteSession(rj *RequestCollection) {

	defer rj.SignalFinish()

	var wg sync.WaitGroup
	retry := make(chan *RequestSend, 10) // Requests for those that need a retry or they have finsihed retrying
	result := rj.Result                  // Scrape results
	// safe := rj.Safe
	cancel := rj.Cancel

	cDone := 0
	cClient := 0
	cHeader := 0

	for _, x := range rj.RS {
		go HandleRequest(x, retry, result, &wg)
	}

loop:
	for {
		select {
		// Handling retries
		case item := <-retry:
			// if item.Caught {
			// 	cHeader = changeHeaders(item.Request.Request, rj.RJ, cHeader)
			// 	go HandleRequest(item, retry, result, &wg)
			// } else if item.Retries == 0 {
			// 	cDone++
			// 	continue
			// } else {
			// 	item.Client = rj.RJ.Clients[cClient]
			// 	cClient++
			// 	if cClient == len(rj.RJ.Clients) {
			// 		cClient = 0
			// 	}
			// 	go HandleRequest(item, retry, result, &wg)
			// }
			switch {
			case item.Caught:
				cHeader = changeHeaders(item.Request.Request, rj.RJ, cHeader)
				go HandleRequest(item, retry, result, &wg)
				break
			case item.Retries == 0:
				cDone++
				break
			default:
				item.Client = rj.RJ.Clients[cClient]
				cClient++
				if cClient == len(rj.RJ.Clients) {
					cClient = 0
				}
				go HandleRequest(item, retry, result, &wg)

			}
			break
		case <-cancel:
			break loop
		default:
			if cDone == len(rj.RJ.Links) {
				break loop
			}
		}
	}

	// This closes goroutine if it is run as a goroutine
	return
}

func Simple(rc *RequestCollection) {
	defer rc.SignalFinish()

	var wg sync.WaitGroup
	// RETRY has NO functionality
	retry := make(chan *RequestSend)
	result := rc.Result
	cancel := rc.Cancel
	finish := make(chan struct{})

	for _, x := range rc.RS {
		go HandleRequest(x, retry, result, &wg)
	}

	go func() {
		wg.Wait()
		finish <- struct{}{}
		return
	}()

loop:
	for {
		select {
		case <-cancel:
			break loop
		case <-finish:
			break loop
		case <-retry:
			break
		}
	}

	return
}

// ----------------------------------------------------------

func HandleRequest(req *RequestSend, retry chan *RequestSend, rr *RequestResult, wg *sync.WaitGroup) {

	var resp *http.Response
	var err error

	wg.Add(1)
	defer wg.Done()

	client := req.Client
	request := req.Request.Request

	if client != nil {
		respX, errX := client.Do(request)
		resp = respX
		err = errX
	} else {
		cli := http.Client{}
		respX, errX := cli.Do(request)
		resp = respX
		err = errX
	}

	if err != nil {
		log.Println("ProxyConnection: Client Failed! [POOL]")
		// resp.Body.Close()
		req.Decrement()
		retry <- req
		return
	}

	fmt.Println("----------------------------Success-------------------------------")
	defer resp.Body.Close()

	// data, err := traverse.HTMLDocUTF8(resp)
	data, err := RunScrape(resp, rr, req.Method)
	if err != nil {
		log.Println("Couldn't read body")
		return
	}

	// we were requesting with no proxy so no new client was needed so we are finished here
	if client == nil {
		return
	}

	// if the request has gone successfull but the scrape has not i.e. website has detected it's being scraped and providing non-essential information
	if !data {
		req.Caught = true
		retry <- req
		return
	}

	// rr.add(data)
	req.Retries = 0
	retry <- req

	return
}

// ----------------------------------------------------------

func changeClient(client *http.Client, list []*http.Client, counter int) int {
	client = list[counter]
	if counter == len(list) {
		counter = 0
	}
	return counter
}

func changeHeaders(req *http.Request, jar *RequestJar, count int) int {
	req.Header = *jar.Headers[count]
	count++
	if count == len(jar.Headers) {
		return 0
	}
	return count
}

// if true, it means that the scrape was unsuccessful
// if false, scrape successful
// it is up toxw the user to add their scraped data into RequestResult
func RunScrape(r *http.Response, res *RequestResult, m func(doc *traverse.HTMLDocument, rr *RequestResult) bool) (bool, error) {
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

	item := traverse.HTMLDocBytes(&bytes)

	return m(&item, res), err
}
