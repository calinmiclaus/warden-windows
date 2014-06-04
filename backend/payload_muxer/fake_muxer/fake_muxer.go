package fake_muxer

import (
	"io"
	"sync"

	"github.com/cloudfoundry-incubator/garden/warden"
)

type FakeMuxer struct {
	source      io.Reader
	subscribers map[uint32][]chan<- warden.ProcessStream

	lock *sync.RWMutex
}

func New() *FakeMuxer {
	return &FakeMuxer{
		lock: new(sync.RWMutex),

		subscribers: make(map[uint32][]chan<- warden.ProcessStream),
	}
}

func (muxer *FakeMuxer) SetSource(source io.Reader) {
	muxer.lock.Lock()
	muxer.source = source
	muxer.lock.Unlock()
}

func (muxer *FakeMuxer) Subscribe(processID uint32, stream chan<- warden.ProcessStream) {
	muxer.lock.Lock()
	muxer.subscribers[processID] = append(muxer.subscribers[processID], stream)
	muxer.lock.Unlock()
}

func (muxer *FakeMuxer) Subscribers(processID uint32) []chan<- warden.ProcessStream {
	muxer.lock.RLock()
	subscribers := make([]chan<- warden.ProcessStream, len(muxer.subscribers))
	copy(subscribers, muxer.subscribers[processID])
	muxer.lock.RUnlock()

	return subscribers
}
