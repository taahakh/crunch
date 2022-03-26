package speed

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func ScrapeSession(rj *RequestCollection, n int) {
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
	rr.Alive = "I SEE THIS AFTERWARDS"

	rj.Cancel = make(chan struct{}, n+1)
	jobs := make(chan *RequestSend, n+1)
	retry := make(chan *RequestSend, n+1)
	rj.Safe = make(chan struct{}, n+1)
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
		fmt.Println(rr.Count())
		rj.Safe <- struct{}{}
		rj.Done = true
	}()

}

func SSWorker(jobs <-chan *RequestSend, retry chan *RequestSend, closeChannel chan struct{}, rr *RequestResult, wg *sync.WaitGroup) {
	for {
		select {
		case item := <-jobs:
			wg.Add(1)
			HandleRequest(item, retry, rr, wg)
		case <-closeChannel:
			return
		}
	}
}

func HandleRequest(req *RequestSend, retry chan *RequestSend, rr *RequestResult, wg *sync.WaitGroup) struct{} {
	defer wg.Done()

	client := req.Client
	request := req.Request

	resp, err := client.Do(request)
	if err != nil {
		log.Println("ProxyConnection: Client Failed! [ScrapeSession]")
		req.Decrement()
		retry <- req
		return struct{}{}
	}
	fmt.Println("----------------------------Success-------------------------------")
	defer resp.Body.Close()

	data, err := HTMLDocUTF8(resp)
	if err != nil {
		log.Println("Couldn't read body")
		return struct{}{}
	}

	rr.add(data)
	req.Retries = 0
	retry <- req

	return struct{}{}
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
