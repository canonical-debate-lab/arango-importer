package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

func main() {
	fmt.Println("Starting data migration")

	db, _ := OpenArangoConnection()

	// Open collections for vertices
	colClaims := openCollection(db, "claims", true)
	colArgs := openCollection(db, "arguments", true)

	filename := "data/Test1.json"
	//filename := "data/small_test.json"
	//filename := "data/single_test.json"
	fmt.Println("Loading file:", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error loading file:", err.Error())
		panic(err.Error())
	}

	// Convert the JSON to nodes
	data := []DebateMapNode{}
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		fmt.Println("Error parsing JSON:", err.Error())
		panic(err.Error())
	}

	// First pass: create Claims and Arguments
	claims := make(map[string]Claim)
	args := make(map[string]Argument)
	mpClaims := []DebateMapNode{}
	idxs := []int{}
	for i, node := range data {
		fmt.Printf("Read node: +%v\n", node)
		switch node.Type {
		case NODE_TYPE_CLAIM:
			claim := NewClaim(node)
			claims[claim.ID] = claim
			createItem(colClaims, claim)
		case NODE_TYPE_ARGUMENT:
			if node.MultiPremise {
				// In Debate Map, it's the Arguments that are MP
				// In this graph, it will be an MP Claim instead, which needs to be created

				// Replace the new node with claim and arg nodes
				argNode, claimNode := node.ConvertToMPClaim()
				data[i] = argNode
				mpClaims = append(mpClaims, claimNode)

				idxs = append(idxs, i)

				claim := NewClaim(claimNode)
				claims[claim.ID] = claim
				createItem(colClaims, claim)

				argument := NewArgument(argNode)
				args[argument.ID] = argument
				createItem(colArgs, argument)
			} else {
				argument := NewArgument(node)
				args[argument.ID] = argument
				createItem(colArgs, argument)
			}
		}
	}
	data = append(data, mpClaims...)

	// Open collections for edges
	edgeInferences := openCollection(db, "inferences", true)
	edgeBaseClaims := openCollection(db, "base_claims", true)
	edgePremises := openCollection(db, "premises", true)

	// Second pass: create edges
	for _, node := range data {
		fmt.Printf("Read item for edges: +%v\n", node)
		switch node.Type {
		case NODE_TYPE_CLAIM:
			if node.MultiPremise {
				for _, childVal := range node.Children {
					child := NewChildFromData(childVal)
					if child != nil {
						if claim, ok := claims[child.ID]; ok {
							createPremise(edgePremises, node.ArangoID(), claim.ArangoID())
						} else {
							panic(fmt.Sprintf("Child Premise %s not found", child.ID))
						}
					}
				}
			} else {
				for _, childVal := range node.Children {
					child := NewChildFromData(childVal)
					if child != nil {
						if arg, ok := args[child.ID]; ok {
							arg.Pro = child.IsPro()
							updateItem(colArgs, arg.ID, arg)
							createInference(edgeInferences, node.ArangoID(), arg.ArangoID())
						} else if claim, ok := claims[child.ID]; ok {
							// Data consistency problem in the Debate Map version!
							// Create an intervening Argument to resolve the problem
							arg := Argument{
								ID:        fmt.Sprintf("ldkfasf%s", child.ID), // TODO - generate UUID and base64 it
								CreatedAt: claim.CreatedAt,
								Creator:   claim.Creator,
								Pro:       child.IsPro(),
							}
							createItem(colArgs, arg)
							createInference(edgeInferences, node.ArangoID(), arg.ArangoID())
							createBaseClaim(edgeBaseClaims, arg.ArangoID(), claim.ArangoID())
						} else {
							panic(fmt.Sprintf("Child Argument %s not found", child.ID))
						}
					}
				}
			}
		case NODE_TYPE_ARGUMENT:
			for _, childVal := range node.Children {
				child := NewChildFromData(childVal)
				if child != nil {
					if arg, ok := args[child.ID]; ok {
						arg.Pro = child.IsPro()
						updateItem(colArgs, arg.ID, arg)
						createInference(edgeInferences, node.ArangoID(), arg.ArangoID())
					} else if claim, ok := claims[child.ID]; ok {
						createBaseClaim(edgeBaseClaims, node.ArangoID(), claim.ArangoID())
					} else {
						panic(fmt.Sprintf("Child %s not found", child.ID))
					}
				}
			}
		}
	}

	fmt.Println("Done.")
}

func OpenArangoConnection() (driver.Database, error) {
	fmt.Println("Connecting to the database")
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		fmt.Println("Error connecting the the database:", err.Error())
		panic(err.Error())
	}
	conn, err = conn.SetAuthentication(driver.BasicAuthentication("root", ""))
	if err != nil {
		fmt.Println("Error setting the connection authentication:", err.Error())
		panic(err.Error())
	}
	c, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})
	if err != nil {
		fmt.Println("Error creating the database client:", err.Error())
		panic(err.Error())
	}

	db, err := c.Database(nil, "canonical_debate")
	if err != nil {
		fmt.Println("Error choosing the database:", err.Error())
		panic(err.Error())
	}

	return db, err
}

func createItem(c driver.Collection, item interface{}) {
	meta, err := c.CreateDocument(nil, item)
	if err != nil {
		fmt.Printf("Error creating item: %s\nItem: %+v\n", err.Error(), item)
		panic(err.Error())
	}
	fmt.Println("Created item. Meta:", meta)
}

func updateItem(c driver.Collection, id string, item interface{}) {
	meta, err := c.UpdateDocument(nil, id, item)
	if err != nil {
		fmt.Printf("Error updating item: %s\nItem: %+v\n", err.Error(), item)
		panic(err.Error())
	}
	fmt.Println("Updated item. Meta:", meta)
}

func createInference(c driver.Collection, fromid, toid string) {
	inference := Inference{
		From: fromid,
		To:   toid,
	}
	createItem(c, inference)
}

func createPremise(c driver.Collection, fromid, toid string) {
	premise := Premise{
		From: fromid,
		To:   toid,
	}
	createItem(c, premise)
}

func createBaseClaim(c driver.Collection, fromid, toid string) {
	bc := BaseClaim{
		From: fromid,
		To:   toid,
	}
	createItem(c, bc)
}

func openCollection(db driver.Database, name string, truncate bool) driver.Collection {
	col, err := db.Collection(nil, name)
	if err != nil {
		fmt.Printf("Error opening %s collection: %s\n", name, err.Error())
		panic(err.Error())
	}
	if truncate {
		err = col.Truncate(nil)
		if err != nil {
			fmt.Printf("Error truncating %s: %s", name, err.Error())
			panic(err.Error())
		}
	}
	return col
}
