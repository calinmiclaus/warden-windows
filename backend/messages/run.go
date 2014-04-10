package messages

type RunRequest struct {
	Script     string `json:"script"`
	Privileged bool   `json:"privileged"`
}

type RunResponse struct {
	ProcessID uint32 `json:"process_id"`
}
