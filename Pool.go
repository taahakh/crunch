package speed

import (
	"fmt"
	"log"
	"sync"
)

type Pool struct {
	mu          sync.RWMutex
	name        string // Pool Name
	collections map[string]*RequestCollection
	end         chan string
}

// Setting an identifier for our pool
func (p *Pool) SetName(name string) {
	p.name = name
	p.collections = make(map[string]*RequestCollection, 0)
	go p.captureCompleted()
}

// Add collection to the map
func (p *Pool) Add(col string, rc *RequestCollection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.collections[col]; !ok {
		rc.Identity = col
		p.collections[col] = rc
	} else {
		log.Println("This collection ID already exists")
	}
}

// Delete collection from the map
func (p *Pool) Remove(col string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	delete(p.collections, col)
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
func (p *Pool) PopIfCompleted(id string) *RequestResult {
	var rr *RequestResult
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.collections[id]; ok {
		if val.Done {
			rr = val.Result
			p.rem(id)
			return rr
		}
	}

	return rr
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

// Run scrape
func (p *Pool) Run(id, method string, n int) {
	switch method {
	case "scrapesession":
		go ScrapeSession(p.collections[id], n)
		break
	}
}

func (p *Pool) captureCompleted() {

}

// ------------------------------------------------------------------------
