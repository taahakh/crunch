package req

import (
	"fmt"
	"sync"
)

// A circular queue that works like a stack
//
// This is mutexed as multiple goroutines might
// access the same structure
type Queue struct {
	mu         sync.RWMutex
	List       []*Send
	front, pop int
}

func (q *Queue) New(length int) *Queue {
	return &Queue{
		List:  make([]*Send, 0, length),
		front: 0,
		pop:   0,
	}
}

// Make assigns items of Send array to queue
func (q *Queue) Make(rs []*Send) *Queue {
	q.List = rs
	q.front = 0
	q.pop = 0
	return q
}

// Add places Send into the queue
//
// It will all to the queue if there is an empty space
// If there is no empty space, the slice will be appended
// with the new element. It will fill nil places when it has
// been popped
func (q *Queue) Add(rs *Send) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	// Check if the current position is empty
	if q.List[q.front] == nil {
		q.List[q.front] = rs
		q.front++
		// check if the next position is ok
		if q.front == len(q.List) {
			q.front = 0
		}
		return true
	} else {
		// We append to the slice if the slice if full
		// Add position will stay the same until it empty and has
		// added item
		q.List = append(q.List, rs)
	}

	return false
}

// Pop removes item from the queue in FIFO order
func (q *Queue) Pop() *Send {
	q.mu.Lock()
	defer q.mu.Unlock()
	// we want to pop a non empty object
	if q.List[q.pop] != nil {
		temp := q.List[q.pop]
		q.List[q.pop] = nil
		q.pop++
		if q.pop == len(q.List) {
			q.pop = 0
		}
		return temp
	}

	return nil
}

// View shows the list at this current state
func (q *Queue) View() {
	q.mu.RLock()
	defer q.mu.RUnlock()
	fmt.Println("List: ", q.List)
}
