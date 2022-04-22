package req

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

type Pool struct {
	mu   sync.RWMutex
	name string // Pool Name

	// Active or soon to be active collections
	collections map[string]*RequestCollection

	/* Pool Usage */

	// Channel to collect all ended collections. Collects collection identifier and is stored in finsihed
	end chan string
	// end chan *RequestCollection

	// Stores finished collections
	finished []string
	// finished []*RequestCollection

	// Signal to close the Garbage collector and the pool. Closing pool will come later
	close chan struct{}
}

// Setting an identifier for our pool
func (p *Pool) SetName(name string) {
	p.name = name
	p.collections = make(map[string]*RequestCollection, 0)
	p.finished = make([]string, 0)
	// p.finished = make([]*RequestCollection, 0)
	p.close = make(chan struct{})
	p.end = make(chan string)
	// p.end = make(chan *RequestCollection)
	go p.collector()
}

// Add collection to the map
func (p *Pool) Add(col string, rc *RequestCollection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.collections[col]; !ok {
		rc.Identity = col
		rc.Notify = &p.end
		// rc.Safe = make(chan struct{})
		rc.Cancel = make(chan struct{})
		rc.Result = &RequestResult{}
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

	return errors.New("Cannot remove until this is collection is finished or cancelled")
}

// Unsafe removal. Should be used when collection is safe to remove
// The collection will run even when removed
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

// Ends all request. Doesn't allow graceful finish
func (p *Pool) CancelCollection(id string) (*RequestResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if val, ok := p.collections[id]; ok {
		if val.Done {
			return val.Result, nil
		}
		val.Done = true
		val.Cancel <- struct{}{}
		for _, x := range val.RS {
			if x == nil {
				continue
			}
			c := *x.Request.Cancel
			c()
		}
		return val.Result, nil
	}

	return nil, errors.New("Collection doesn't exist")
}

//  Removes successfully when the requestcollection has been given done status
func (p *Pool) PopIfCompleted(id string) (*RequestResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.collections[id]; ok {
		if val.Done {
			rr := val.Result
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
	case "complete":
		go CompleteSession(p.collections[id])
		break
	case "batch":
		go Batch(p.collections[id], 2, "3s")
		break
	}
}

// Garbage collector for the pool
func (p *Pool) collector() {
	for {
		select {
		case x := <-p.end:
			p.mu.Lock()
			p.finished = append(p.finished, x)

			// p.finished = append(p.finished, x)
			// delete(p.collections, x.Identity)
			p.mu.Unlock()
			break
		case <-p.close:
			fmt.Println("Im meant to close GCGCGGCGCCGG")
			return
		}
	}
}

func (p *Pool) closeGC() {
	p.close <- struct{}{}
}

func (p *Pool) GracefulClose() {

}

func (p *Pool) Close() {
	for _, x := range p.collections {
		p.CancelCollection(x.Identity)
	}
	fmt.Println("sdfjfijdskfjdjk")
	// p.closeGC()
	return
}

// ------------------------------------------------------------------------
