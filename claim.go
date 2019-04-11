package main

import (
	"fmt"
	"time"
)

type Claim struct {
	ID            string    `json:"_key"`
	CreatedAt     time.Time `json:"start"`
	Creator       string    `json:"creator"`
	Title         string    `json:"title"`
	Negation      string    `json:"negation"`
	Question      string    `json:"question"`
	Note          string    `json:"note"`
	MultiPremise  bool      `json:"mp"`
	ChildrenOrder []string  `json:"childOrder"`
}

func (claim Claim) ArangoID() string {
	return fmt.Sprintf("claims/%s", claim.ID)
}

func NewClaim(node DebateMapNode) Claim {
	return Claim{
		ID:            node.ID,
		CreatedAt:     node.CreatedTime(),
		Creator:       node.Creator,
		Title:         node.Current.Title.Base,
		Negation:      node.Current.Title.Negation,
		Question:      node.Current.Title.Question,
		Note:          node.Note,
		MultiPremise:  node.MultiPremise,
		ChildrenOrder: node.ChildrenOrder,
	}
}
