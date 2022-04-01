package speed

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Global sleep, timeouts, batched workers
// func ScrapeSession(rj *RequestCollection, n int) {
// 	var t time.Duration
// 	var setSleep bool

// 	if rj.Finish == UntilComplete {
// 		setSleep = false
// 	} else {
// 		var err error

// 		t, err = time.ParseDuration(rj.Finish)
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}

// 		setSleep = true
// 	}

// 	var wg sync.WaitGroup
// 	var rr RequestResult

// 	rj.Result = &rr

// 	rj.Cancel = make(chan struct{}, n+1)  // Pool/Collection to close this collection
// 	jobs := make(chan *RequestSend, n+1)  // Send jobs to complete
// 	retry := make(chan *RequestSend, n+1) // Requests that need to be handled again but with a different client
// 	rj.Safe = make(chan struct{}, 1)      // Tells Pool that it is safe to access and all requests have been handled
// 	closeChannel := rj.Cancel

// 	for i := 0; i < n; i++ {
// 		go SSWorker(jobs, retry, closeChannel, &rr, &wg)
// 	}

// 	for _, x := range rj.RS {
// 		jobs <- x
// 	}

// 	cClients := 0
// 	cDone := 0

// 	go func() {
// 		if setSleep {
// 			time.Sleep(t)
// 			fmt.Println("Finished Sleeping")
// 			closeChannel <- struct{}{}
// 		}
// 		return
// 	}()

// loop:
// 	for {
// 		select {
// 		case item := <-retry:
// 			if item.Retries == 0 {
// 				cDone++
// 				continue
// 			} else {
// 				item.Client = rj.RJ.Clients[cClients]
// 				cClients++
// 				if cClients == len(rj.RJ.Clients) {
// 					cClients = 0
// 				}
// 				jobs <- item
// 			}
// 			break
// 		case <-closeChannel:
// 			break loop
// 		default:
// 			if cDone == len(rj.RJ.Links) {
// 				closeChannel <- struct{}{}
// 				break loop
// 			}
// 		}
// 	}

// 	go func() {
// 		wg.Wait()
// 		fmt.Println("Done waiting?")
// 		close(closeChannel)
// 		rj.Done = true
// 		rj.Safe <- struct{}{}
// 		*rj.Notify <- rj.Identity
// 		return
// 	}()
// }

func ScrapeSession(rj *RequestCollection, n int) {
	// Setting global timeout or not
	var t time.Duration
	var setSleep bool

	if rj.Finish == UntilComplete {
		setSleep = false
	} else {
		var err error

		t, err = time.ParseDuration(rj.Finish)
		if err != nil {
			log.Println(err)
			return
		}

		setSleep = true
	}

	var wg sync.WaitGroup
	var rr RequestResult

	rj.Result = &rr

	rj.Cancel = make(chan struct{}, n+1)  // Pool/Collection to close this collection
	jobs := make(chan *RequestSend, n+1)  // Send jobs to complete
	retry := make(chan *RequestSend, n+1) // Requests that need to be handled again but with a different client
	rj.Safe = make(chan struct{}, 1)      // Tells Pool that it is safe to access and all requests have been handled
	closeChannel := rj.Cancel

	for i := 0; i < n; i++ {
		go SSWorker(jobs, retry, closeChannel, &rr, &wg)
	}

	for _, x := range rj.RS {
		jobs <- x
	}

	cClients := 0
	cDone := 0

	go func() {
		if setSleep {
			time.Sleep(t)
			fmt.Println("Finished Sleeping")
			closeChannel <- struct{}{}
		}
		return
	}()

loop:
	for {
		select {
		case item := <-retry:
			if item.Retries == 0 {
				cDone++
				continue
			} else {
				item.Client = rj.RJ.Clients[cClients]
				cClients++
				if cClients == len(rj.RJ.Clients) {
					cClients = 0
				}
				jobs <- item
			}
			break
		case <-closeChannel:
			break loop
		default:
			if cDone == len(rj.RJ.Links) {
				closeChannel <- struct{}{}
				break loop
			}
		}
	}

	go func() {
		wg.Wait()
		fmt.Println("Done waiting?")
		close(closeChannel)
		rj.Done = true
		rj.Safe <- struct{}{}
		*rj.Notify <- rj.Identity
		return
	}()
}

func SSWorker(jobs <-chan *RequestSend, retry chan *RequestSend, closeChannel chan struct{}, rr *RequestResult, wg *sync.WaitGroup) {
	for {
		select {
		case item := <-jobs:
			// wg.Add(1)
			HandleRequest(item, retry, rr, wg)
		case <-closeChannel:
			return
		}
	}
}

// ---------------------------------------------------------------------------------

// Handles retries, individual request timeouts and cancellations,
// all request cancellations

func CompleteSession(rj *RequestCollection) {

	defer rj.SignalFinish()

	var wg sync.WaitGroup
	retry := make(chan *RequestSend, 10) // Requests for those that need a retry or they have finsihed retrying
	result := rj.Result                  // Scrape results
	safe := rj.Safe
	// cancel := rj.Cancel

	cDone := 0
	cClient := 0

	for _, x := range rj.RS {
		go HandleRequest(x, retry, result, &wg)
	}

loop:
	for {
		select {
		// Handling retries
		case item := <-retry:
			if item.Retries == 0 {
				cDone++
				continue
			} else {
				item.Client = rj.RJ.Clients[cClient]
				cClient++
				if cClient == len(rj.RJ.Clients) {
					cClient = 0
				}
				go HandleRequest(item, retry, result, &wg)
			}
			break
		case <-safe:
			// case <-cancel:
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

// func changeClient(rs *RequestSend, client []*http.Client, c int) *RequestSend {
// 	rs.Client = client[c]
// 	return rs
// }

func HandleRequest(req *RequestSend, retry chan *RequestSend, rr *RequestResult, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	client := req.Client
	request := req.Request

	resp, err := client.Do(request)
	if err != nil {
		log.Println("ProxyConnection: Client Failed! [POOL]")
		req.Decrement()
		retry <- req
		return
	}

	fmt.Println("----------------------------Success-------------------------------")
	defer resp.Body.Close()

	data, err := HTMLDocUTF8(resp)
	if err != nil {
		log.Println("Couldn't read body")
		return
	}

	rr.add(data)
	req.Retries = 0
	retry <- req

	return
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
