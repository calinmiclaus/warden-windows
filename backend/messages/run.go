package messages

type RunRequest struct {
	Script     string `json:"kill"`
	Privileged bool   `json:"privileged"`
}

type RunResponse struct {
	ProcessID uint32 `json:"process_id"`
}
