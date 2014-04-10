package payload_muxer

import (
	"encoding/json"
	"io"
	"log"
	"sync"

	"github.com/cloudfoundry-incubator/warden-windows/backend/messages"

	"github.com/cloudfoundry-incubator/garden/backend"
)

type Muxer interface {
	SetSource(io.Reader)
	Subscribe(processID uint32, stream chan<- backend.ProcessStream)
}

type PayloadMuxer struct {
	subscribers     map[uint32][]chan<- backend.ProcessStream
	subscribersLock *sync.Mutex
}

func New() PayloadMuxer {
	return PayloadMuxer{
		subscribers:     make(map[uint32][]chan<- backend.ProcessStream),
		subscribersLock: new(sync.Mutex),
	}
}

func (muxer PayloadMuxer) SetSource(stream io.Reader) {
	go muxer.dispatch(stream)
}

func (muxer PayloadMuxer) Subscribe(processID uint32, stream chan<- backend.ProcessStream) {
	muxer.subscribersLock.Lock()
	muxer.subscribers[processID] = append(muxer.subscribers[processID], stream)
	muxer.subscribersLock.Unlock()
}

func (muxer PayloadMuxer) dispatch(stream io.Reader) {
	decoder := json.NewDecoder(stream)

	var payload messages.ProcessPayload

	for {
		err := decoder.Decode(&payload)
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Println("decode error:", err)
			return
		}

		muxer.subscribersLock.Lock()

		subscribers := muxer.subscribers[payload.ProcessID]

		if payload.ExitStatus != nil {
			stream := backend.ProcessStream{
				ExitStatus: payload.ExitStatus,
			}

			for _, sub := range subscribers {
				select {
				case sub <- stream:
				default:
				}

				close(sub)
			}

			delete(muxer.subscribers, payload.ProcessID)
		} else {
			stream := backend.ProcessStream{
				Source: payload.Source,
				Data:   payload.Data,
			}

			for _, sub := range subscribers {
				select {
				case sub <- stream:
				default:
				}
			}
		}

		muxer.subscribersLock.Unlock()
	}
}
