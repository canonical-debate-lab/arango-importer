package main

import (
	"time"

	"github.com/google/uuid"
)

const NODE_TYPE_CATEGORY int = 10
const NODE_TYPE_PACKAGE int = 20
const NODE_TYPE_QUESTION = 30
const NODE_TYPE_CLAIM int = 40
const NODE_TYPE_ARGUMENT int = 50

const ARGUMENT_POLARITY_PRO int = 10
const ARGUMENT_POLARITY_CON int = 20

const ARGUMENT_TYPE_ANY int = 10
const ARGUMENT_TYPE_ANY_TWO int = 15
const ARGUMENT_TYPE_ALL int = 20

type DebateMapNode struct {
	ID            string                 `json:"_key"`
	CreatedAt     int64                  `json:"createdAt"`
	Creator       string                 `json:"creator"`
	Type          int                    `json:"type"`
	Current       Current                `json:"current"`
	Note          string                 `json:"note"`
	Polarity      int                    `json:"polarity"`
	MultiPremise  bool                   `json:"multiPremiseArgument"`
	Parents       map[string]interface{} `json:"parents"`
	Children      map[string]interface{} `json:"children"`
	ChildrenOrder []string               `json:"childrenOrder"`
}

func (node DebateMapNode) CreatedTime() time.Time {
	return time.Unix(0, node.CreatedAt*1000000)
}

func (node DebateMapNode) IsPro() bool {
	return node.Polarity == ARGUMENT_POLARITY_PRO
}

func (node DebateMapNode) ChildOrder(childId string) int {
	for i, id := range node.ChildrenOrder {
		if childId == id {
			return i + 1
		}
	}
	return 0
}

// Creates a new MP Claim node,
// And changes current node to point to it as its base claim
func (node DebateMapNode) ConvertToMPClaim() (newArg, newClaim DebateMapNode) {
	argChildren := map[string]interface{}{}
	claimChildren := map[string]interface{}{}
	for _, childInt := range node.Children {
		if child := NewChildFromData(childInt); child != nil {
			if child.Polarity > 0 {
				argChildren[child.ID] = *child
			} else {
				claimChildren[child.ID] = *child
			}
		}
	}

	newArg = DebateMapNode{
		ID:           node.ID,
		CreatedAt:    node.CreatedAt,
		Creator:      node.Creator,
		Type:         NODE_TYPE_ARGUMENT,
		Polarity:     node.Polarity,
		MultiPremise: false,
		Parents:      node.Parents,
		Children:     argChildren,
	}

	newClaim = DebateMapNode{
		ID:            uuid.New().String(),
		CreatedAt:     node.CreatedAt,
		Creator:       node.Creator,
		Type:          NODE_TYPE_CLAIM,
		Current:       node.Current,
		Note:          node.Note,
		Parents:       map[string]interface{}{newArg.ID: newArg},
		Children:      claimChildren,
		MultiPremise:  true,
		ChildrenOrder: node.ChildrenOrder,
	}

	newArg.Children[newClaim.ID] = Child{ID: newClaim.ID}

	return
}

type Current struct {
	Title        TitleSet `json:"titles"`
	ArgumentType int      `json:"argumentType"`
}

type TitleSet struct {
	Base     string `json:"base"`
	Negation string `json:"negation"`
	Question string `json:"yesNoQuestion"`
}

type Child struct {
	ID       string `json:"_key"`
	Polarity int    `json:"polarity"`
}

func (child Child) IsPro() bool {
	return child.Polarity == ARGUMENT_POLARITY_PRO
}

func NewChildFromData(data interface{}) *Child {
	if m, ok := data.(map[string]interface{}); ok {
		if id, okay := m["_key"].(string); okay {
			child := Child{ID: id}
			if polarity, kk := m["polarity"].(float64); kk {
				child.Polarity = int(polarity)
			}
			return &child
		}
	} else if ch, ok := data.(Child); ok {
		return &ch
	}

	return nil
}

/* Claim
+map[_key:zsrQ9ZRGSg2y1QDg0y_Xxg children:map[Kp8pR1UyRpC-5SI6sdd_VA:map[_:true _key:Kp8pR1UyRpC-5SI6sdd_VA polarity:10] _key:children] createdAt:1.542072143141e+12 creator:fG4HB6nP5baRQwZZ6BjrLuSOjjD2 current:map[_key:8RJXTx1ZT0yrnAp_bZqzJw accessLevel:10 createdAt:1.542072143305e+12 creator:fG4HB6nP5baRQwZZ6BjrLuSOjjD2 node:zsrQ9ZRGSg2y1QDg0y_Xxg titles:map[_key:titles allTerms:map[_key:allTerms change:true climate:true far:true fighting:true has:true impact:true in:true investment:true little:true roi:true shown:true so:true the:true very:true] base:The investment in fighting climate change so far has shown very little impact (ROI).]] currentRevision:8RJXTx1ZT0yrnAp_bZqzJw parents:map[_key:parents wvwbFY_1Rx2qqEL819X0aw:true] type:40]
*/

/* Multi-premise Argument
{"children":{"1Pl8F_cmT-W84XrF1rvgaA":{"_":true,"form":10,"_key":"1Pl8F_cmT-W84XrF1rvgaA"},"wTVYg4c-QLmI7QjLjcjckw":{"_":true,"_key":"wTVYg4c-QLmI7QjLjcjckw"},"_key":"children"},"childrenOrder":["wTVYg4c-QLmI7QjLjcjckw","1Pl8F_cmT-W84XrF1rvgaA"],"createdAt":1551183882923,"creator":"fG4HB6nP5baRQwZZ6BjrLuSOjjD2","currentRevision":"zSGFJw44Sm2M6zFNP8ti2g","multiPremiseArgument":true,"parents":{"kwsLLiNFSTmbokQ1_nO-bA":{"_":true,"_key":"kwsLLiNFSTmbokQ1_nO-bA"},"_key":"parents"},"type":50,"_key":"Ikan0wFzSXm7GYSPvglJ3A","current":{"accessLevel":10,"argumentType":20,"createdAt":1551386119594,"creator":"fG4HB6nP5baRQwZZ6BjrLuSOjjD2","node":"Ikan0wFzSXm7GYSPvglJ3A","titles":{"allTerms":{"a":true,"brasil":true,"deveria":true,"dos":true,"esperar":true,"executivo":true,"fazer":true,"governo":true,"militares":true,"n":true,"o":true,"object":true,"para":true,"protecionista":true,"quando":true,"reforma":true,"respeito":true,"tiver":true,"um":true,"_key":"allTerms"},"base":"O Brasil deveria esperar para fazer a reforma para quando não tiver um governo executivo protecionista a respeito dos militares.","_key":"titles"},"_key":"zSGFJw44Sm2M6zFNP8ti2g"}}
*/

/* Argument
{"children":{"0IuVkaiSSqeDIUYoqJmHgg":{"_":true,"_key":"0IuVkaiSSqeDIUYoqJmHgg"},"_key":"children"},"childrenOrder":["0IuVkaiSSqeDIUYoqJmHgg"],"createdAt":1552339936133,"creator":"VoJg7aKCtgWj3SR4Ailk4VWxcYv2","currentRevision":"lVt1SFoiQTCjXl6_adEy0Q","parents":{"unqeHLthRwuFuVrwP3PeRQ":{"_":true,"_key":"unqeHLthRwuFuVrwP3PeRQ"},"_key":"parents"},"type":50,"_key":"m5TYIVrtQqicm0AsuEnn7Q","current":{"accessLevel":10,"argumentType":20,"createdAt":1552339936134,"creator":"VoJg7aKCtgWj3SR4Ailk4VWxcYv2","node":"m5TYIVrtQqicm0AsuEnn7Q","titles":{"allTerms":{"_key":"allTerms"},"base":"","_key":"titles"},"_key":"lVt1SFoiQTCjXl6_adEy0Q"}}
*/
