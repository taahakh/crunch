package req

import (
	"sync"
	"time"
)

// Batch sends out requests in grouped batches.
// size - how many requests per batch
// gap - how long until to send next batch. if the time gap = 0, then we will return. there needs to be a timed batch
// THIS DOES WORK BUT IN A LIMITED FASHION - THERE ARE KNOWN ERRORS
func Batch(rc *Collection, size int, gap string, handler BatchHandler) {

	// Prevent usage of Batch
	return
	var dur time.Duration
	var handle func(item *Send, wg *sync.WaitGroup, q *Queue, rc *Collection)
	var done func() bool

	// check if time duration is empty
	if gap != "" {
		t, err := time.ParseDuration(gap)
		if err != nil {
			return
		}
		dur = t
	}

	if dur == 0 {
		return
	}

	var wg sync.WaitGroup
	retry := make(chan *Send, size)
	end := make(chan struct{})

	collectionChecker(rc)

	result := rc.Result
	cancel := rc.Cancel
	complete := rc.Complete
	ms := rc.muxrs
	ms.SetChannel(retry)
	ms.Manual()

	if handler != nil {
		handle = handler.Handle
		done = handler.Done
	} else {
		dh := &DefaultBatchHandler{done: len(rc.RS)}
		handle = dh.Handle
		done = dh.Done
	}

	q := Queue{}
	q.Make(rc.RS)

	wg.Add(1)

	go func() {
		// We go thorugh the list first without doing the retries
		for {
			q.View()
			select {
			case <-end:
				return
			default:
				for i := 0; i < size; i++ {
					item := q.Pop()
					if item == nil {
						continue
					}
					wg.Add(1)
					go HandleRequest(true, item, retry, result, ms, &wg)
				}

				time.Sleep(dur)
			}

		}
	}()

	go func() {

		for {
			time.Sleep(time.Millisecond * sleepTime)

			select {
			case item := <-retry:
				handle(item, &wg, &q, rc)
				break
			case <-cancel:
				end <- struct{}{}
				wg.Done()
				return
			default:
				if done() {
					end <- struct{}{}
					wg.Done()
					return
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		// cancel <- struct{}{}
		if complete != nil {
			complete <- rc.Identity
		}
		rc.Done = true
		return
	}()

	return
}
