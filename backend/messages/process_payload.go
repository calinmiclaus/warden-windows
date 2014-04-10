package messages

import (
	"github.com/cloudfoundry-incubator/garden/backend"
)

type ProcessPayload struct {
	ProcessID  uint32                      `json:"process_id"`
	Source     backend.ProcessStreamSource `json:"source"`
	Data       []byte                      `json:"data"`
	ExitStatus *uint32                     `json:"exit_status"`
}
