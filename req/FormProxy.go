package req

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/taahakh/speed/traverse"
	"golang.org/x/net/html/charset"
)

type er struct {
	text string
}

// Proxy constants
const (
	ProxyConnectionPoolError = "ProxyConnection: Client Failed! [POOL]"
	sleepTime                = 1000
)

// ---------------------------------------------------------------------------------

// Handles retries, individual request timeouts and cancellations,
// all request cancellations

// request - number of request to push per batch

// size - how many requests per batch
// gap - how long until to send next batch. if the time gap = 0, then we will return. there needs to be a timed batch
func Batch(rc *Collection, size int, gap string) {
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
	retry := make(chan *Send, size)
	// Scrape results
	result := rc.Result
	// safe := rj.Safe
	// The pool telling the collection to end all requests
	cancel := rc.Cancel
	// End the goroutine where we do requests.
	// This goroutine ends when the second goroutine receives a cancellation singal from the pool
	// sends struct to end this goroutine. Cannot use the same cancellation signal to end the goroutine
	// we also want to make sure that any retries that are updated to the queue does not get missed by the cancellation
	end := make(chan struct{})
	complete := rc.Complete

	cClient := 0
	// cDone := len(rj.RJ.Links)
	cDone := len(rc.RS)
	cHeader := 0

	ms := rc.muxrs
	ms.SetChannel(retry)

	q := Queue{}
	q.Make(rc.RS)

	wg.Add(1)

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
					wg.Add(1)
					go HandleRequest(true, item, retry, result, ms, &wg)
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
				if item.Caught {
					cHeader = changeHeaders(item.Request.Request, rc.RJ, cHeader)
					item.Caught = false
					q.Add(item)
				} else if item.Retries == 0 {
					cDone--
				} else {
					cClient = changeClient(item.Client, rc.RJ.Clients, cClient)
					q.Add(item)
				}
				break
			case <-cancel:
				end <- struct{}{}
				wg.Done()
				return
			default:
				if cDone == 0 {
					end <- struct{}{}
					wg.Done()
					return
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		if complete != nil {
			*complete <- rc.Identity
		}
		rc.Done = true
		return
	}()

	return
}

// ----------------------------------------------------------

func CompleteSession(rc *Collection) {

	var wg sync.WaitGroup
	retry := make(chan *Send, 10) // Requests for those that need a retry or they have finsihed retrying
	end := make(chan struct{})

	cClient := 0
	cHeader := 0

	// ms := rc.muxrs
	// if ms != nil {
	// 	ms.SetChannel(retry)
	// }

	collectionChecker(rc)
	result := rc.Result     // Scrape results
	cancel := rc.Cancel     // Pool usage
	complete := rc.Complete // Pool usage
	ms := rc.muxrs
	ms.SetChannel(retry)

	for _, x := range rc.RS {
		wg.Add(1)
		go HandleRequest(true, x, retry, result, ms, &wg)
	}

	go func() {
		wg.Wait()
		if complete != nil {
			*complete <- rc.Identity
		}
		close(end)
		rc.Done = true
		return
	}()

loop:
	for {
		select {
		// Handling retries
		case item := <-retry:
			switch {
			case item.Caught:
				wg.Add(1)
				cHeader = changeHeaders(item.Request.Request, rc.RJ, cHeader)
				item.Caught = false
				go HandleRequest(true, item, retry, result, ms, &wg)
				break
			case item.Retries == 0:
				break
			default:
				wg.Add(1)
				cClient = changeClient(item.Client, rc.RJ.Clients, cClient)
				go HandleRequest(true, item, retry, result, ms, &wg)
			}
			break
		case <-cancel:
			break loop
		case <-end:
			break loop
		}
	}

	// completeCriterion(retry, rj, result, &cancel, &end, &wg)

	// This closes goroutine if it is run as a goroutine
	return
}

// ----------------------------------------------------------

// Simple handles requests normally. It can take in proxied clients or not and doesn't
// allow retries.
func Simple(rc *Collection) {
	var wg sync.WaitGroup

	// Retry functionality is set to continue scraping new calls and NOT failed calls
	// However its can be setup in any way
	retry := make(chan *Send)
	finish := make(chan struct{})

	// ms := rc.muxrs
	// if ms != nil {
	// 	ms.SetChannel(retry)
	// }

	collectionChecker(rc)
	result := rc.Result
	cancel := rc.Cancel
	complete := rc.Complete
	ms := rc.muxrs
	ms.SetChannel(retry)

	fmt.Println(result, cancel, complete)

	for _, x := range rc.RS {
		wg.Add(1)
		go HandleRequest(false, x, retry, result, ms, &wg)
	}

	go func() {
		wg.Wait()
		close(finish)
		if complete != nil {
			*complete <- rc.Identity
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

func clientProcess(client *http.Client, request *http.Request, req *Send, retry chan *Send) (*http.Response, error) {
	fmt.Println(request)
	resp, err := client.Do(request)

	if err != nil {
		req.Decrement()
		retry <- req
		return nil, errors.New(ProxyConnectionPoolError)
	}

	return resp, nil
}

func noClientProcess(client *http.Client, request *http.Request) (*http.Response, error) {
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

func HandleRequest(enforce bool, req *Send, retry chan *Send, rr *Store, ms *MutexSend, wg *sync.WaitGroup) {

	var resp *http.Response
	var err error

	// wg.Add(1)
	if wg != nil {
		defer wg.Done()
	}

	client := req.Client
	request := req.Request.Request

	// fmt.Println(client.Transport)

	if enforce {
		resp, err = clientProcess(client, request, req, retry)
	} else {
		resp, err = noClientProcess(client, request)
	}

	// REMEMBER TO REMOVE
	rr.Add(er{text: "Nice"})

	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("----------------------------Success-------------------------------")
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

	// rr.add(data)
	req.Retries = 0
	retry <- req

	return
}

// ----------------------------------------------------------

func changeClient(client *http.Client, list []*http.Client, counter int) int {
	client = list[counter]
	counter++
	if counter == len(list) {
		counter = 0
	}
	return counter
}

func changeHeaders(req *http.Request, jar *Jar, count int) int {
	req.Header = *jar.Headers[count]
	count++
	if count == len(jar.Headers) {
		return 0
	}
	return count
}

// if true, it means that the scrape was unsuccessful
// if false, scrape successful

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

	item := traverse.HTMLDocBytes(&bytes)
	pack := Result{}
	pack = pack.New(&item, res, ms)
	if m == nil {
		return true, err
	}
	return m(pack), err
}

func collectionChecker(rc *Collection) *Collection {
	if rc.Cancel == nil {
		rc.Cancel = make(chan struct{}, 1)
	}

	// if rc.Notify == nil {
	// 	temp := make(chan string, 1)
	// 	rc.Notify = &temp
	// }

	// if rc.Complete == nil {
	// 	temp := make(chan string, 1)
	// 	rc.Complete = &temp
	// }

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
