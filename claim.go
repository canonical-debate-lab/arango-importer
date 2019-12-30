package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const PREMISE_RULE_NONE int = 0
const PREMISE_RULE_ALL int = 1
const PREMISE_RULE_ANY int = 2
const PREMISE_RULE_ANY_TWO int = 3

type Claim struct {
	Key          string    `json:"_key"`
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"start"`
	Creator      string    `json:"creator"`
	Title        string    `json:"title"`
	Negation     string    `json:"negation"`
	Question     string    `json:"question"`
	Note         string    `json:"note"`
	MultiPremise bool      `json:"mp"`
	PremiseRule  int       `json:"mprule"`
	Truth        float32   `json:"truth"`
}

func (claim Claim) ArangoID() string {
	return fmt.Sprintf("claims/%s", claim.Key)
}

func NewClaim(node DebateMapNode) Claim {
	return Claim{
		Key:          uuid.New().String(),
		ID:           node.ID,
		CreatedAt:    node.CreatedTime(),
		Creator:      node.Creator,
		Title:        node.Current.Title.Base,
		Negation:     node.Current.Title.Negation,
		Question:     node.Current.Title.Question,
		Note:         node.Note,
		MultiPremise: node.MultiPremise,
		PremiseRule:  argumentTypeToPremiseRule(node.Current.ArgumentType),
		Truth:        0.50,
	}
}

func argumentTypeToPremiseRule(argumentType int) int {
	switch argumentType {
	case ARGUMENT_TYPE_ANY:
		return PREMISE_RULE_ANY
	case ARGUMENT_TYPE_ANY_TWO:
		return PREMISE_RULE_ANY_TWO
	case ARGUMENT_TYPE_ALL:
		return PREMISE_RULE_ALL
	default:
		return PREMISE_RULE_NONE
	}
}
