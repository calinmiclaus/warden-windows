package fake_muxer

import (
	"io"
	"sync"

	"github.com/cloudfoundry-incubator/garden/backend"
)

type FakeMuxer struct {
	source      io.Reader
	subscribers map[uint32][]chan<- backend.ProcessStream

	lock *sync.RWMutex
}

func New() *FakeMuxer {
	return &FakeMuxer{
		lock: new(sync.RWMutex),

		subscribers: make(map[uint32][]chan<- backend.ProcessStream),
	}
}

func (muxer *FakeMuxer) SetSource(source io.Reader) {
	muxer.lock.Lock()
	muxer.source = source
	muxer.lock.Unlock()
}

func (muxer *FakeMuxer) Subscribe(processID uint32, stream chan<- backend.ProcessStream) {
	muxer.lock.Lock()
	muxer.subscribers[processID] = append(muxer.subscribers[processID], stream)
	muxer.lock.Unlock()
}

func (muxer *FakeMuxer) Subscribers(processID uint32) []chan<- backend.ProcessStream {
	muxer.lock.RLock()
	subscribers := make([]chan<- backend.ProcessStream, len(muxer.subscribers))
	copy(subscribers, muxer.subscribers[processID])
	muxer.lock.RUnlock()

	return subscribers
}
