package main

// A Premise is an edge that goes from a Multi-premise Claim
// to one of the Claims that represents a specific premise
type Premise struct {
	From string `json:"_from,omitempty"`
	To   string `json:"_to,omitempty"`
}
