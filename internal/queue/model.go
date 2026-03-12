package queue

// JobPayload is the contract between the API and the Worker
type JobPayload struct {
	ID      string `json:"id"`
	Type    string `json:"type"`    // e.g., "EMAIL", "IMAGE_RESIZE"
	Payload string `json:"payload"` // Flexible data (could be a JSON string)
}
