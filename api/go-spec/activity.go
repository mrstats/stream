package activity

import (
	"time"

	"github.com/google/uuid"
)

type Request struct {
	Actor     string    `json:"actor"`
	Verb      string    `json:"verb"`
	Object    string    `json:"object"`
	ForeignID string    `json:"foreign_id,omitempty"`
	Target    string    `json:"target,omitempty"`
	Time      time.Time `json:"time,string,omitempty"`
	To        []string  `json:"to,omitempty"`
}

type Response struct {
	ID uuid.UUID `json:"id,string"`
	Request
}

type ActivitiesResponse struct {
	Activities []Response `json:"activities"`
}
