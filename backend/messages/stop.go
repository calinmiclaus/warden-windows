package messages

type StopRequest struct {
	Kill bool `json:"kill"`
}

type StopResponse struct{}
