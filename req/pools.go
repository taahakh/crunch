package req

import (
	"errors"
	"log"
	"sync"
	"time"
)

// PoolLook is mainly to interface Pool struct
// Used to prevent other usage with the pool
type PoolLook interface {

	// Closes pool
	Close()

	// Returns the collections once all collections are finished
	Collections() map[string]*Collection
}

// PoolSettings handles what to do when collections are all finished
// and what to do when completed collection are just finished
type PoolSettings struct {
	// Runs when all collections have been completed
	AllCollectionsCompleted func(p PoolLook)

	// Runs when the current finished collection is completed
	IncomingCompletedCollections func(rc *Collection)
}

// RequestMethods are constants on request/scraping methods
type RequestMethods int

// For RequestMethods
const (
	Method_Complete RequestMethods = iota
	Method_Batch
	Method_Simple
)

// Pool holds multiple collections to be run at whatever time.
// Requests for collections can be closed and access to the collections via
// the pool can be made for further usage e.g. via a web server
type Pool struct {
	mu sync.RWMutex

	// name for the pool
	name string

	// collections map
	collections map[string]*Collection

	/* --------------- Pool Usage --------------- */

	// complete SIGNALS when a gorotuine has finished
	// Collects collection identifier and is stored in finsihed
	complete chan string

	// finished collects collections names. We do not store Collection as functionality would
	// require dereferencing when it doesn't have to be
	finished map[string]struct{}

	// close the Garbage collector and the pool.
	close chan struct{}

	// closed prevents deadlock if we try to close again even though it's already closed
	closed bool
}

// New instantiates and sets an identifier for our pool
func (p *Pool) New(name string, settings PoolSettings) {
	p.name = name
	p.collections = make(map[string]*Collection, 0)
	p.finished = make(map[string]struct{}, 0)
	// We buffer the close channel as the collector can be used to close itself via the close method
	// This however makes it so that the collector wont be able to get the close signal while there is
	// a stuck process in the close method (cant reach the collector).
	// Running Close() as another goroutine can be another way but its not recommeneded as their can be a
	// loss of data/race conditions when doing so
	p.close = make(chan struct{}, 1)
	// p.complete = make(chan string, 1)
	p.complete = make(chan string)
	go p.collector(settings)
}

// Add collection to the map
func (p *Pool) Add(col string, rc *Collection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.collections[col]; !ok {
		rc.Identity = col
		rc.Complete = p.complete
		rc.Cancel = make(chan struct{})
		rc.Result = &Store{
			mu:      sync.Mutex{},
			res:     make([]interface{}, 0),
			counter: 0,
		}
		rc.SetMuxSend()
		p.collections[col] = rc
	} else {
		log.Println("This collection ID already exists")
	}
}

// Remove collection from the map
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

// rem
// UNSAFE - Deprecitated
// Should be used when collection is safe to remove
// The collection will run even when removed
func (p *Pool) rem(id string) {
	delete(p.collections, id)
}

// Refresh UNSAFE
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

// CancelCollection Ends all request. Doesn't allow graceful finish
func (p *Pool) CancelCollection(id string) (*Store, error) {
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

// PopIfCompleted Pops from collection and returns the scraped data
// ------ RENAME
func (p *Pool) ReturnIfCompleted(id string) ([]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.isFinished(id) {
		val, _ := p.collections[id]
		result := val.Result.Read()
		return result, nil
	}

	return nil, errors.New("Could not pop. Either it was not completed or the id is wrong")
}

// isFinished checks if the collections have safely closed and are completed
func (p *Pool) isFinished(id string) bool {
	return p.amIDone(id) && p.amIFinished(id)
}

// BlockUntilComplete will block the current function until the collection has been completed
func (p *Pool) BlockUntilComplete(id string) []interface{} {
	for {
		res, err := p.ReturnIfCompleted(id)
		if err == nil {
			return res
		}
		time.Sleep(time.Millisecond * 100)
	}
}

// NumOfRunnig sees how many Collections are running
// EXPENSIVE
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

// Checks if finished wtthout MUTEX.
func (p *Pool) amIFinished(id string) bool {
	_, ok := p.finished[id]
	return ok
}

// AmIFinished checks if collections are finished
func (p *Pool) AmIFinished(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.amIFinished(id)
}

// Checks if collection is done without mutex
func (p *Pool) amIDone(id string) bool {
	if col, ok := p.collections[id]; ok {
		if col.Done {
			return true
		}
		return false
	}
	return false
}

// AmIDone checks if the collection is done
func (p *Pool) AmIDone(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.amIDone(id)
}

// Run scrape
// NEED TO CHANGE
func (p *Pool) Run(id string, method RequestMethods, n int) {
	p.collections[id].Start = true
	switch method {
	case Method_Complete:
		go Session(p.collections[id], nil)
		break
	case Method_Batch:
		// Known errors with batch
		// go Batch(p.collections[id], 2, "3s", nil)
		break
	case Method_Simple:
		go Run(p.collections[id])
		break
	}
}

func (p *Pool) RunSimple(id string) {
	p.collections[id].Start = true
	go Run(p.collections[id])
}

func (p *Pool) RunSession(id string, handler CompleteHandler) {
	p.collections[id].Start = true
	go Session(p.collections[id], handler)
}

// Garbage collector for the pool
func (p *Pool) collector(settings PoolSettings) {
	for {
		select {
		case y := <-p.complete:
			// if we want to do something when the collection has finished with the scraped data
			p.mu.Lock()
			p.finished[y] = struct{}{}
			if settings.IncomingCompletedCollections != nil {
				settings.IncomingCompletedCollections(p.collections[y])
			}
			if len(p.finished) == len(p.collections) {
				if settings.AllCollectionsCompleted != nil {
					p.mu.Unlock()
					settings.AllCollectionsCompleted(p)
					p.mu.Lock()
				}
			}
			p.mu.Unlock()
			break
		case <-p.close:
			return
		}
	}
}

// Collections returns collection map
func (p *Pool) Collections() map[string]*Collection {
	return p.collections
}

// Prints how many stored collections there are
func (p *Pool) Stored() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.collections)
}

// Stop ends all running collections
func (p *Pool) Stop() {
	for _, x := range p.collections {
		p.CancelCollection(x.Identity)
	}
}

// CompletionChecker
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

	if count == len(p.finished) {
		return true
	}

	return false
}

// Close ends collector and closes the pool
func (p *Pool) Close() {

	if p.closed {
		return
	}

	p.Stop()

	for !p.CompletionChecker() {
		time.Sleep(time.Second * 1)
	}

	// Closing collector
	p.close <- struct{}{}

	p.closed = true
	return
}

// ------------------------------------------------------------------------
