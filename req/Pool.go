package req

import (
	"errors"
	"log"
	"sync"
	"time"
)

type PoolLook interface {
	Close()
	Collections() map[string]*RequestCollection
}

type PoolSettings struct {
	AllCollectionsCompleted      func(p PoolLook)
	IncomingCompletedCollections func(rc *RequestCollection)
	IncomingRequestCompletion    func(name string)
}

type RequestMethods int

const (
	Method_Complete RequestMethods = iota
	Method_Batch
	Method_Simple
)

type Pool struct {
	mu   sync.RWMutex
	name string // Pool Name

	// Active or soon to be active collections
	collections map[string]*RequestCollection

	/* Pool Usage */

	// SIGNALS where a gorotuine has finished
	// Collects collection identifier and is stored in finsihed
	complete chan string

	// Stores finished collections names. We do not store requestcollection as functionality would
	// require dereferencing when it doesn't have to be
	finished map[string]struct{}

	// Signal to close the Garbage collector and the pool. Closing pool will come later
	close chan struct{}

	// Create deadlock if we try to close again even though it's already closed
	closed bool
}

// Setting an identifier for our pool
func (p *Pool) New(name string, settings PoolSettings) {
	p.name = name
	p.collections = make(map[string]*RequestCollection, 0)
	p.finished = make(map[string]struct{}, 0)
	p.close = make(chan struct{})
	// p.complete = make(chan string, 1)
	p.complete = make(chan string)
	go p.collector(settings)
}

// Add collection to the map
func (p *Pool) Add(col string, rc *RequestCollection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.collections[col]; !ok {
		rc.Identity = col
		rc.Complete = &p.complete
		rc.Cancel = make(chan struct{})
		rc.Result = &RequestResult{}
		rc.SetMuxSend()
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

// UNSAFE
// Should be used when collection is safe to remove
// The collection will run even when removed
func (p *Pool) rem(id string) {
	delete(p.collections, id)
}

// UNSAFE
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
		// fmt.Println("before: ", val.Done)
		val.Done = true
		// fmt.Println("before: ", val.Done)
		val.Cancel <- struct{}{}

		// Stops requests from occuring via mutexed RequestSend
		// Saving of information will still occur
		go val.GetMuxSend().End()

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

//  Pops from collection and returns the scraped data
func (p *Pool) PopIfCompleted(id string) ([]*interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.collections[id]; ok {
		rr, err := p.popCompleted(val, id)
		result := rr.Read()
		p.rem(id)
		return result, err
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

// EXPENSIVE
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

// Return how many collections have been completed
func (p *Pool) NumOfCompleted() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.finished)
}

// Returns list for the collections that have finished
func (p *Pool) GetFinishedList() map[string]struct{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.finished
}

func (p *Pool) AmIFinished(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.finished[id]
	return ok
}

func (p *Pool) AmIDone(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if col, ok := p.collections[id]; ok {
		if col.Done {
			return true
		}
		return false
	}
	return false
}

// Run scrape
func (p *Pool) Run(id string, method RequestMethods, n int) {
	p.collections[id].Start = true
	switch method {
	case Method_Complete:
		go CompleteSession(p.collections[id])
		break
	case Method_Batch:
		go Batch(p.collections[id], 2, "3s")
		break
	case Method_Simple:
		go Simple(p.collections[id])
		break
	}
}

// Garbage collector for the pool
func (p *Pool) collector(settings PoolSettings) {
	for {
		select {
		case y := <-p.complete:
			// fmt.Println("COMPLETED: ", y)
			// pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
			// if we want to do something when the collection has finished with the scraped data
			p.mu.Lock()
			p.finished[y] = struct{}{}
			if settings.IncomingCompletedCollections != nil {
				settings.IncomingCompletedCollections(p.collections[y])
				// removes collection from the pool. The name will still exist in finished array
				// but gone for good
				// p.rem(y)
			}
			p.mu.Unlock()
			break
		case <-p.close:
			return
			// default:
			// 	break
		}
	}
}

func (p *Pool) Collections() map[string]*RequestCollection {
	return p.collections
}

// Prints how many stored collections there are
func (p *Pool) Stored() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.collections)
}

func (p *Pool) Stop() {
	for _, x := range p.collections {
		p.CancelCollection(x.Identity)
	}
}

// UNSAFE - implemented safely
func (p *Pool) CompletionChecker() bool {
	// p.mu.Lock()
	// defer p.mu.Unlock()
	count := 0
	for _, x := range p.collections {
		if x.Done {
			count++
		}
	}

	// fmt.Println("Count: ", count)
	// fmt.Println("Finished: ", len(p.finished))

	if count == len(p.finished) {
		return true
	}

	return false
}

func (p *Pool) Close() {

	if p.closed {
		return
	}

	p.Stop()

	for !p.CompletionChecker() {
		time.Sleep(time.Second * 1)
	}

	// fmt.Println("Finished checking -> equls ")

	// Closing collector
	p.close <- struct{}{}

	p.closed = true
	return
}

// ------------------------------------------------------------------------
