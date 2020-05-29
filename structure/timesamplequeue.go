package structure

import "sync"

// TimedSampleQueue stores received samples and keeping their scheduled times
type TimedSampleQueue struct {
	buffer []timedSample
	head   int
	tail   int

	cond      *sync.Cond
	headMutex sync.RWMutex
	tailMutex sync.RWMutex
}

// Add a timed sample at the end of the queue
func (q *TimedSampleQueue) Add(sample [2]float64, time int64) {
	q.headMutex.Lock()
	defer q.headMutex.Unlock()

	q.waitNotFull()

	q.buffer[q.head%len(q.buffer)] = timedSample{sample: sample, time: time}
	q.head = q.inc(q.head)

	q.cond.Broadcast()
}

// Remove the first timed sample from the queue and return it
func (q *TimedSampleQueue) Remove() (sample [2]float64, time int64) {
	q.tailMutex.Lock()
	defer q.tailMutex.Unlock()

	q.waitNotEmpty()

	v := q.buffer[q.tail%len(q.buffer)]
	q.tail = q.inc(q.tail)

	q.cond.Broadcast()
	return v.sample, v.time
}

// Peek the first timed sample without removing it from the queue
func (q *TimedSampleQueue) Peek() (sample [2]float64, time int64) {
	q.tailMutex.RLock()
	defer q.tailMutex.RUnlock()

	q.waitNotEmpty()

	v := q.buffer[q.tail%len(q.buffer)]
	return v.sample, v.time
}

// Length (size) of the queue
func (q *TimedSampleQueue) Length() int {
	q.tailMutex.RLock()
	defer q.tailMutex.RUnlock()
	q.headMutex.RLock()
	defer q.headMutex.RUnlock()
	if q.tail <= q.head {
		return q.head - q.tail
	}
	return q.head - q.tail + 2*len(q.buffer)
}

// Capacity of the queue
func (q *TimedSampleQueue) Capacity() int {
	return len(q.buffer)
}

func (q *TimedSampleQueue) inc(i int) int {
	return (i + 1) % (2 * len(q.buffer))
}

func (q *TimedSampleQueue) full() bool {
	return (q.tail+len(q.buffer))%(2*len(q.buffer)) == q.head
}

func (q *TimedSampleQueue) empty() bool {
	return q.head == q.tail
}

func (q *TimedSampleQueue) waitNotFull() {
	q.cond.L.Lock()
	for q.full() {
		q.cond.Wait()
	}
	q.cond.L.Unlock()
}

func (q *TimedSampleQueue) waitNotEmpty() {
	q.cond.L.Lock()
	for q.empty() {
		q.cond.Wait()
	}
	q.cond.L.Unlock()
}

// NewTimedSampleQueue creates a new queue of specified size
func NewTimedSampleQueue(size int) *TimedSampleQueue {
	return &TimedSampleQueue{buffer: make([]timedSample, size), head: 0, tail: 0, cond: sync.NewCond(&sync.Mutex{})}
}

type timedSample struct {
	sample [2]float64
	time   int64
}
