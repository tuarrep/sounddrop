package structure

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/tuarrep/sounddrop/message"
	"time"
)

type queueable struct {
	popAt  time.Time
	sample *message.StreamData
}

// TimedSampleQueue buffers queueables and emits them at exact playing time
type TimedSampleQueue struct {
	queueables      []*queueable
	subscribers     map[int]chan *message.StreamData
	subscriberCount int
	started         bool
}

// Push adds a sample to queue
func (tsq *TimedSampleQueue) Push(sample *message.StreamData, popAt time.Time) int {
	tsq.queueables = append(tsq.queueables, &queueable{sample: sample, popAt: popAt})

	return len(tsq.queueables)
}

// Subscribe adds subscriber to send it queueables
func (tsq *TimedSampleQueue) Subscribe(subscriber chan *message.StreamData) int {
	tsq.subscribers[tsq.subscriberCount] = subscriber

	sid := tsq.subscriberCount
	tsq.subscriberCount++

	return sid
}

// Unsubscribe unregisters a listener
func (tsq *TimedSampleQueue) Unsubscribe(sid int) {
	delete(tsq.subscribers, sid)
}

// Start emitting queueables
func (tsq *TimedSampleQueue) Start() {
	if tsq.started {
		return
	}

	tsq.started = true
	go tsq.loop()
}

func (tsq *TimedSampleQueue) loop() {
	for {
		pop, next := tsq.pop(false)

		if pop {
			nextTime, _ := ptypes.Timestamp(next.NextAt)
			if nextTime.Before(time.Now().Add(1*time.Millisecond)) && nextTime.After(time.Now().Add(-9*time.Millisecond)) {
				for _, subscriber := range tsq.subscribers {
					subscriber <- next
				}

				_, _ = tsq.pop(true)
			} else if nextTime.Before(time.Now().Add(-6 * time.Millisecond)) {
				// Too late for this one
				_, _ = tsq.pop(true)
				continue
			}
		}

		if pop, streamData := tsq.pop(false); pop {
			nextTime, _ := ptypes.Timestamp(streamData.NextAt)
			time.Sleep(time.Until(nextTime.Add(-8 * time.Millisecond)))
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (tsq *TimedSampleQueue) pop(remove bool) (bool, *message.StreamData) {
	if len(tsq.queueables) == 0 {
		return false, nil
	}

	queueable := tsq.queueables[0]

	if remove {
		tsq.queueables = tsq.queueables[1:]
	}

	return true, queueable.sample
}

// NewTimedSampleQueue creates an new TimedSampleQueue
func NewTimedSampleQueue() *TimedSampleQueue {
	return &TimedSampleQueue{queueables: make([]*queueable, 0), subscribers: map[int]chan *message.StreamData{}, subscriberCount: 0, started: false}
}
