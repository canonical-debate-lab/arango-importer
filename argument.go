package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Argument struct {
	Key              string    `json:"_key"`
	ID               string    `json:"id"`
	CreatedAt        time.Time `json:"start"`
	Creator          string    `json:"creator"`
	TargetClaimID    *string   `json:"targetClaimId,omitempty"`
	TargetArgumentID *string   `json:"targetArgId,omitempty"`
	ClaimID          string    `json:"claimId"`
	Title            string    `json:"title"`
	Negation         string    `json:"negation"`
	Question         string    `json:"question"`
	Note             string    `json:"note"`
	Pro              bool      `json:"pro"`
}

func (arg Argument) ArangoID() string {
	return fmt.Sprintf("arguments/%s", arg.Key)
}

func NewArgument(node DebateMapNode) Argument {
	return Argument{
		Key:       uuid.New().String(),
		ID:        node.ID,
		CreatedAt: node.CreatedTime(),
		Creator:   node.Creator,
		Title:     node.Current.Title.Base,
		Negation:  node.Current.Title.Negation,
		Question:  node.Current.Title.Question,
		Note:      node.Note,
	}
}
