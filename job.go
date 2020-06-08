package background

import (
	"net/http"
	"time"
)

// Status represents status of background job.
type Status string

// List of available statuses of background jobs.
const (
	StatusWorking   = Status("WORKING")
	StatusCompleted = Status("COMPLETED")
)

// Job represents background job.
type Job struct {
	ID        string    `json:"id"`
	Status    Status    `json:"status"`
	Response  *Response `json:"response"`
	Kind      string    `json:"kind"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Response represents response of origin handler which runs in background.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       string
}
