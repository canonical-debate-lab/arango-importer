package main

import (
	"time"
)

// A Premise is an edge that goes from a Multi-premise Claim
// to one of the Claims that represents a specific premise
type Premise struct {
	Key       string    `json:"_key"`
	CreatedAt time.Time `json:"start"`
	Creator   string    `json:"creator"`
	From      string    `json:"_from,omitempty"`
	To        string    `json:"_to,omitempty"`
	Order     int       `json:"order"`
}
