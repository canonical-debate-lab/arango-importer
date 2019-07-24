package main

import (
	"time"
)

// Inference is an edge from the target (a Claim or Argument) of an Argument
// to the Argument that is making the inference
type Inference struct {
	Key       string    `json:"_key"`
	CreatedAt time.Time `json:"start"`
	Creator   string    `json:"creator"`
	From      string    `json:"_from,omitempty"`
	To        string    `json:"_to,omitempty"`
}
