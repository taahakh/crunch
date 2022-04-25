package req

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/taahakh/speed/traverse"
)

// ---------------------------------------------------------------------------------

// Handles retries, individual request timeouts and cancellations,
// all request cancellations

// request - number of request to push per batch
// gap - how long to wait until sending next batch

// size - how many requests per batch
// gap - how long until to send next batch. if the time gap = 0, then each batch is done sequentially
func Batch(rj *RequestCollection, size int, gap string) {
	defer rj.SignalFinish()

	var dur time.Duration

	if gap != "" {
		t, err := time.ParseDuration(gap)
		if err != nil {
			return
		}
		dur = t
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

	q := Queue{}
	q.Make(rj.RS)

	go func(req []*RequestSend) {
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
	}(rj.RS)

	go func() {
		for {

			time.Sleep(time.Millisecond * 1000)

			select {
			case item := <-retry:
				if item.Caught == true {
					changeUserAgent(item.Request.Request)
					q.Add(item)
				} else if item.Retries == 0 {
					cDone--
				} else {
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

	for _, x := range rj.RS {
		go HandleRequest(x, retry, result, &wg)
	}

loop:
	for {
		select {
		// Handling retries
		case item := <-retry:
			if item.Caught == true {
				changeUserAgent(item.Request.Request)
				go HandleRequest(item, retry, result, &wg)
			} else if item.Retries == 0 {
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
		case <-cancel:
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

func HandleRequest(req *RequestSend, retry chan *RequestSend, rr *RequestResult, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	client := req.Client
	request := req.Request.Request

	resp, err := client.Do(request)
	if err != nil {
		log.Println("ProxyConnection: Client Failed! [POOL]")
		req.Decrement()
		retry <- req
		return
	}

	fmt.Println("----------------------------Success-------------------------------")
	defer resp.Body.Close()

	data, err := traverse.HTMLDocUTF8(resp)
	if err != nil {
		log.Println("Couldn't read body")
		return
	}

	rr.add(data)
	req.Retries = 0
	retry <- req

	return
}

func changeClient(client *http.Client, list []*http.Client, counter int) int {
	client = list[counter]
	if counter == len(list) {
		counter = 0
	}
	return counter
}

func changeUserAgent(req *http.Request) {

}
