package main

import (
	"fmt"
	"time"
)

type Argument struct {
	ID        string    `json:"_key"`
	CreatedAt time.Time `json:"start"`
	Creator   string    `json:"creator"`
	Title     string    `json:"title"`
	Negation  string    `json:"negation"`
	Question  string    `json:"question"`
	Note      string    `json:"note"`
	Pro       bool      `json:"pro"`
}

func (arg Argument) ArangoID() string {
	return fmt.Sprintf("arguments/%s", arg.ID)
}

func NewArgument(node DebateMapNode) Argument {
	return Argument{
		ID:        node.ID,
		CreatedAt: node.CreatedTime(),
		Creator:   node.Creator,
		Title:     node.Current.Title.Base,
		Negation:  node.Current.Title.Negation,
		Question:  node.Current.Title.Question,
		Note:      node.Note,
	}
}
