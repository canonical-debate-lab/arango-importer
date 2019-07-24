package main

import (
	"time"
)

// BaseClaim is an edge pointing from an Argument to the Claim on which it is based
// (the true/false part of the Argument)
type BaseClaim struct {
	Key       string    `json:"_key"`
	CreatedAt time.Time `json:"start"`
	Creator   string    `json:"creator"`
	From      string    `json:"_from,omitempty"`
	To        string    `json:"_to,omitempty"`
}
