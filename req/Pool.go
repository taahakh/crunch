package req

import (
	"errors"
	"log"
	"sync"
	"time"
)

type RequestMethods int

const (
	METHOD_COMPLETE RequestMethods = iota
	METHOD_BATCH
	METHOD_SIMPLE
)

type Pool struct {
	mu   sync.RWMutex
	name string // Pool Name

	// Active or soon to be active collections
	collections map[string]*RequestCollection

	/* Pool Usage */

	// Channel to collect all ended collections. Collects collection identifier and is stored in finsihed
	end chan string

	// Stores finished collections names. We do not store requestcollection as functionality would
	// require dereferencing when it doesn't have to be
	// finished []string
	finished map[string]struct{}

	// Signal to close the Garbage collector and the pool. Closing pool will come later
	close chan struct{}
}

// Setting an identifier for our pool
func (p *Pool) SetName(name string, method func(rc *RequestCollection)) {
	p.name = name
	p.collections = make(map[string]*RequestCollection, 0)
	// p.finished = make([]string, 0)
	p.finished = make(map[string]struct{}, 0)
	p.close = make(chan struct{})
	p.end = make(chan string)
	go p.collector(method)
}

// Add collection to the map
func (p *Pool) Add(col string, rc *RequestCollection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.collections[col]; !ok {
		rc.Identity = col
		rc.Notify = &p.end
		rc.Cancel = make(chan struct{})
		rc.Result = &RequestResult{}
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
		// collection started?
		if !val.Start {
			return nil, errors.New("This collection has not started")
		}
		// collection already completed?
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
		rr, err := p.popCompleted(val, id)
		return rr, err
	}

	return nil, errors.New("Could not pop. Either it was not completed or the id is wrong")
}

func (p *Pool) popCompleted(val *RequestCollection, id string) (*RequestResult, error) {

	if val.Done {
		rr := val.Result
		p.rem(id)
		return rr, nil
	}

	return nil, errors.New("Could not pop")
}

// blocking call
func (p *Pool) PopWhenCompleted(id string) (*RequestResult, error) {
	if val, ok := p.collections[id]; ok {
		for {
			rr, err := p.popCompleted(val, id)
			if err == nil {
				return rr, nil
			}
			time.Sleep(time.Second * 3)
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
		if x.Start {
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

// // Returns list for the collections that have finished
// func (p *Pool) GetFinishedList() []string {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()
// 	return p.finished
// }

// Returns list for the collections that have finished
func (p *Pool) GetFinishedList() map[string]struct{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.finished
}

func (p *Pool) AmIFinished(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.collections[id]
	return ok
}

// Run scrape
func (p *Pool) Run(id string, method RequestMethods, n int) {
	p.collections[id].Start = true
	switch method {
	case METHOD_COMPLETE:
		go CompleteSession(p.collections[id])
		break
	case METHOD_BATCH:
		go Batch(p.collections[id], 2, "3s")
		break
	case METHOD_SIMPLE:
		go Simple(p.collections[id])
		break
	}
}

// Garbage collector for the pool
func (p *Pool) collector(method func(rc *RequestCollection)) {
	for {
		select {
		case x := <-p.end:
			p.mu.Lock()
			// p.finished = append(p.finished, x)
			p.finished[x] = struct{}{}
			// if we want to do something when the collection has finished with the scraped data
			if method != nil {
				method(p.collections[x])
				// removes collection from the pool. The name will still exist in finished array
				p.rem(x)
			}
			p.mu.Unlock()
			break
		case <-p.close:
			return
		}
	}
}

// func (p *Pool) closeGC() {
// 	p.close <- struct{}{}
// }

func (p *Pool) Close() {
	// fmt.Println("CLOSING: 5 Second Grace period")
	// time.Sleep(time.Second * 10)
	for _, x := range p.collections {
		p.CancelCollection(x.Identity)
	}
	time.Sleep(time.Millisecond * 500)
	// p.closeGC()
	// Closing collector
	p.close <- struct{}{}
	return
}

// ------------------------------------------------------------------------
