package req

import (
	"sync"
)

type WorkerPool struct {
	mu sync.RWMutex

	name string

	workers       <-chan *RequestSend
	queuedWorkers <-chan *RequestSend

	// Syncs and allow closures of all workers all workers/handlers
	cancelGroup sync.WaitGroup

	// Sends all failed scrapes/request to the retry handler(s)
	retry chan *RequestSend

	/* Retry information */

	jar    *RequestJar
	result *RequestResult
	muxrs  *MutexSend
}

func (wp *WorkerPool) New(name string) {
	wp.mu = sync.RWMutex{}
	wp.name = name
	wp.retry = make(chan *RequestSend)
}

func (wp *WorkerPool) retryHandler() {

}

func (wp *WorkerPool) AddRetryHandler() {

}

func (wp *WorkerPool) AddQueue() {

}

func (wp *WorkerPool) AddWorker() {

}
