package req

import (
	"fmt"
	"sync"
)

// A circular queue
type Queue struct {
	mu         sync.RWMutex
	List       []*RequestSend
	front, pop int
}

func (q *Queue) New(length int) *Queue {
	return &Queue{
		List:  make([]*RequestSend, 0, length),
		front: 0,
		pop:   0,
	}
}

func (q *Queue) Make(rs []*RequestSend) *Queue {
	q.List = rs
	q.front = 0
	q.pop = 0
	return q
}

func (q *Queue) Add(rs *RequestSend) bool {
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
	}

	return false
}

func (q *Queue) Pop() *RequestSend {
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

func (q *Queue) View() {
	q.mu.RLock()
	defer q.mu.RUnlock()
	fmt.Println("List: ", q.List)
}
