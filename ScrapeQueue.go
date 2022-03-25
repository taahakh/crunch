package speed

import (
	"sync"
)

type Pool struct {
	mu          sync.RWMutex
	name        string
	collections map[string]*RequestCollection
}

// Setting an identifier for our pool
func (p *Pool) SetName(name string) {
	p.name = name
}

// Add collection to the map
func (p *Pool) Add(col string, rc *RequestCollection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.collections[col]; !ok {
		p.collections[col] = rc
	}
}

// Delete collection from the map
func (p *Pool) Remove(col string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	delete(p.collections, col)
}

func (p *Pool) safeRemove(id string) {
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

// Cancelling all collections. This doesn't end the pool
func (p *Pool) CancelAll() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for x := range p.collections {
		p.collections[x].Cancel <- struct{}{}
	}
}

// Cancelling a single collection
func (p *Pool) Cancel(id string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.collections[id]; ok {
		val.Cancel <- struct{}{}
	}
}

//  Removes successfully when the requestcollection has been given done status
func (p *Pool) PopIfCompleted(id string) *RequestResult {
	var rr *RequestResult
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.collections[id]; ok {
		rr = val.Result
		p.safeRemove(id)
		return rr
	}

	return rr
}

// ------------------------------------------------------------------------
