package speed

import (
	"fmt"
	"log"
	"sync"
)

func ScrapeSession(rj *RequestCollection, n int) {
	var wg sync.WaitGroup
	var rr RequestResult
	jobs := make(chan *RequestSend, 5)
	retry := make(chan *RequestSend, 2)

	for i := 0; i < n; i++ {
		go SSWorker(jobs, retry, &rr, &wg)
	}

	for _, x := range rj.RS {
		jobs <- x
	}

	cClients := 0
	cDone := 0

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
		}

		if cDone == len(rj.RJ.Links) {
			break
		}
	}

	go func() {
		wg.Wait()
		close(jobs)
		fmt.Println("Goroutine other closing")
		return
		// close()
	}()

	fmt.Println("Im here??")

}

func SSWorker(jobs <-chan *RequestSend, retry chan *RequestSend, rr *RequestResult, wg *sync.WaitGroup) {
	for x := range jobs {
		wg.Add(1)
		HandleRequest(x, retry, rr, wg)
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
