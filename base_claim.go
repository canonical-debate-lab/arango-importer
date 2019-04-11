package main

// BaseClaim is an edge pointing from an Argument to the Claim on which it is based
// (the true/false part of the Argument)
type BaseClaim struct {
	From string `json:"_from,omitempty"`
	To   string `json:"_to,omitempty"`
}
