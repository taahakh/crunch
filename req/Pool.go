package req

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Pool struct {
	mu          sync.RWMutex
	name        string // Pool Name
	collections map[string]*RequestCollection
	/* Pool Usage */

	// Channel to collect all ended collections. Collects collection identifier and is stored in finsihed
	end chan string

	// Stores all finished collections
	finished []string

	// Signal to close the Garbage collector and the pool. Closing pool will come later
	close chan struct{}
}

// Setting an identifier for our pool
func (p *Pool) SetName(name string) {
	p.name = name
	p.collections = make(map[string]*RequestCollection, 0)
	p.finished = make([]string, 0)
	p.close = make(chan struct{})
	p.end = make(chan string)
	go p.captureCompleted()
}

// Add collection to the map
func (p *Pool) Add(col string, rc *RequestCollection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.collections[col]; !ok {
		rc.Identity = col
		rc.Notify = &p.end
		rc.Safe = make(chan struct{})
		// rc.Extend = make(chan *RequestSend)
		p.collections[col] = rc
	} else {
		log.Println("This collection ID already exists")
	}
}

// Delete collection from the map
func (p *Pool) Remove(col string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.collections[col]; ok {
		if val.Done {
			delete(p.collections, col)
			return nil
		}
	}

	return errors.New("Cannot remove until this is collection is finished")
}

// Unsafe removal. Should be used when collection is safe to remove
func (p *Pool) rem(id string) {
	delete(p.collections, id)
}

// Checking our collection and deleting any one that has been finished
func (p *Pool) Refresh() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for x := range p.collections {
		if p.collections[x].Done {
			delete(p.collections, x)
		}
	}
}

// Cancelling does cancel the requests. This WILL and NEEDS to be handled as well
// For now it cancels the workers and any remaining requests are handled and send through

// Cancelling all collections. This doesn't end the pool. This method is NOT safe
func (p *Pool) CancelAll() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for x := range p.collections {
		p.collections[x].Cancel <- struct{}{}
	}
}

// Cancelling a single collection. This is a safe method and makes retrieving results safe as well
func (p *Pool) Cancel(id string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.collections[id]; ok {
		val.Cancel <- struct{}{}
		// Blocking call
		<-val.Safe
		fmt.Println("After blocking call")
	}
}

//  Removes successfully when the requestcollection has been given done status
func (p *Pool) PopIfCompleted(id string) (*RequestResult, error) {
	var rr *RequestResult
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.collections[id]; ok {
		if val.Done {
			rr = val.Result
			p.rem(id)
			return rr, nil
		}
	}

	return nil, errors.New("Could not pop. Either it was not completed or the id is wrong")
}

// See how many Collections are running
func (p *Pool) NumOfRunning() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	counter := 0
	for _, x := range p.collections {
		if x.Done {
			counter++
		}
	}
	return counter
}

// Return how many operations have been completed
func (p *Pool) Completed() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.finished)
}

// Returns list for the collections that have finished
func (p *Pool) GetFinished() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.finished
}

// Run scrape
func (p *Pool) Run(id, method string, n int) {
	switch method {
	case "flood":
		go Flood(p.collections[id], n)
		break
	case "complete":
		go CompleteSession(p.collections[id])
		break
	case "batch":
		go Batch(p.collections[id], 2, "3s")
		break
	}
}

// Garbage collector for the pool
func (p *Pool) captureCompleted() {
	for {
		select {
		case x := <-p.end:
			p.mu.Lock()
			p.finished = append(p.finished, x)
			p.mu.Unlock()
			break
		case <-p.close:
			fmt.Println("Im meant to close GCGCGGCGCCGG")
			return
		}
	}
}

// Ends all request. Doesn't allow graceful finish
func (p *Pool) ForceCancelAll(id string) {
	// fmt.Println("Cancelling")
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.collections[id]; ok {
		val.Safe <- struct{}{}
		for _, x := range val.RS {
			if x == nil {
				continue
			}
			c := *x.Cancel
			c()
		}
	}
}

func (p *Pool) SendForceCancel(id string) {

}

func (p *Pool) Extend(id string, req *http.Request) {
	for val, ok := p.collections[id]; ok; {
		val.RJ.Links = append(val.RJ.Links, req)
		// val.Extend <- req
	}
}

func (p *Pool) CloseGC() {
	p.close <- struct{}{}
}

func (p *Pool) GracefulClose() {

}

func (p *Pool) Close() {
	p.CloseGC()
}

// ------------------------------------------------------------------------
