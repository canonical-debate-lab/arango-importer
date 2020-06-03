package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/google/uuid"
)

const DEFAULT_FILENAME = "data/Test1.json"
const DEFAULT_SERVER = "http://localhost:8529"
const DEFAULT_DB = "canonical_debate"
const DEFAULT_USERNAME = "root"
const DEFAULT_PASSWORD = ""

const FORMAT_UNKNOWN int = 0
const FORMAT_NODES int = 1
const FORMAT_GENERAL int = 2

func main() {
	fmt.Println("Starting data migration")

	var filename, server, dbname, username, password string
	flag.StringVar(&filename, "f", DEFAULT_FILENAME, "filename")
	flag.StringVar(&server, "h", DEFAULT_SERVER, "host (e.g. http://localhost:8529)")
	flag.StringVar(&dbname, "db", DEFAULT_DB, "DB name")
	flag.StringVar(&username, "u", DEFAULT_USERNAME, "username")
	flag.StringVar(&password, "p", DEFAULT_PASSWORD, "password")
	//filename := "data/Test1.json"
	//filename := "data/small_test.json"
	//filename := "data/single_test.json"
	flag.Parse()

	db, _ := OpenArangoConnection(server, dbname, username, password)

	// Open collections for vertices
	colClaims := openCollection(db, "claims", true)
	colArgs := openCollection(db, "arguments", true)

	fmt.Println("Loading file:", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error loading file:", err.Error())
		panic(err.Error())
	}

	var format int
	if strings.HasPrefix(string(file), `[{"children":{`) {
		fmt.Println("---------------------Detected data in NODES format")
		format = FORMAT_NODES
	} else if strings.HasPrefix(string(file), `{"general":`) {
		fmt.Println("---------------------Detected data in GENERAL format")
		format = FORMAT_GENERAL
	} else {
		fmt.Println("---------------------Data is in unknown format")
		format = FORMAT_UNKNOWN
	}

	/*
		m := map[string]interface{}{}
		json.Unmarshal(file, &m)

		for k, _ := range m {
			fmt.Printf("Root key: %+v\n", k)
		}

		data := m["maps"].([]interface{})
		for _, node := range data {
			mnode := node.(map[string]interface{})
			//delete(mnode, "_key")
			fmt.Printf("Read mnode: %+v\n", mnode)
		}
	*/

	// Convert the JSON to nodes
	data := []DebateMapNode{}
	revisions := map[string]NodeRevision{}
	maps := map[string]DebateMapMap{}
	if format == FORMAT_GENERAL {
		general := DebateMapRoot{}
		err = json.Unmarshal(file, &general)
		if err != nil {
			fmt.Println("Error parsing JSON:", err.Error())
			panic(err.Error())
		}
		data = general.Nodes
		revs := general.NodeRevisions
		for _, rev := range revs {
			revisions[rev.ID] = rev
		}
		for _, dmm := range general.Maps {
			maps[dmm.RootNode] = dmm
		}
	} else {
		err = json.Unmarshal(file, &data)
		if err != nil {
			fmt.Println("Error parsing JSON:", err.Error())
			panic(err.Error())
		}
	}

	// First pass: create Claims and Arguments
	claims := make(map[string]Claim)
	args := make(map[string]Argument)
	newClaims := []DebateMapNode{}
	for i, node := range data {
		if node.ID == "" {
			node.ID = node.Current.ID
			data[i] = node
		}
		if format == FORMAT_GENERAL {
			rev := revisions[node.CurrentRevision]
			node.Current.Title = rev.Title
		}
		fmt.Printf("Read node: %+v\n", node)
		switch node.Type {
		case NODE_TYPE_CLAIM:
			claim := NewClaim(node)
			claims[claim.ID] = claim
			createItem(colClaims, claim)
		case NODE_TYPE_ARGUMENT:
			if node.MultiPremise {
				// In Debate Map, it's the Arguments that are MP
				// In this graph, it will be an MP Claim instead, which needs to be created
				if node.ID == "L0Wv33MFQiuWVbWEKcELsA" {
					fmt.Println("----------------------------L0Wv33MFQiuWVbWEKcELsA is MPClaim")
				}

				// Replace the new node with claim and arg nodes
				argNode, claimNode := node.ConvertToMPClaim()
				data[i] = argNode
				newClaims = append(newClaims, claimNode)
				if node.ID == "L0Wv33MFQiuWVbWEKcELsA" {
					fmt.Println("----------------------------For MPClaim L0Wv33MFQiuWVbWEKcELsA created claimNode", claimNode.ID)
					fmt.Println("----------------------------For MPClaim L0Wv33MFQiuWVbWEKcELsA created argNode", argNode.ID)
				}

				claim := NewClaim(claimNode)
				claims[claim.ID] = claim
				createItem(colClaims, claim)

				argument := NewArgument(argNode)
				argument.ClaimID = claim.ID
				args[argument.ID] = argument
				createItem(colArgs, argument)
			} else {
				argument := NewArgument(node)
				args[argument.ID] = argument
				if argument.ID == "L0Wv33MFQiuWVbWEKcELsA" {
					fmt.Println("----------------------------Added L0Wv33MFQiuWVbWEKcELsA to args")
				}
				createItem(colArgs, argument)
			}
		case NODE_TYPE_CATEGORY, NODE_TYPE_PACKAGE, NODE_TYPE_QUESTION:
			// Just to capture node information, these "debate" placeholders will be converted into
			// a claim and (if there's a parent node) an argument
			// They will require manual curation later to make them match the CD concepts
			if dmm, ok := maps[node.ID]; ok {
				node.Current.Title.Base = dmm.Name
			}
			argNode, claimNode := node.ConvertToClaimAndArg()
			if node.ID == "L0Wv33MFQiuWVbWEKcELsA" {
				fmt.Println("----------------------------L0Wv33MFQiuWVbWEKcELsA is a category, package or question")
				fmt.Println("----------------------------For L0Wv33MFQiuWVbWEKcELsA created claimNode", claimNode.ID)
			}

			data[i] = claimNode

			claim := NewClaim(claimNode)
			claims[claim.ID] = claim
			createItem(colClaims, claim)

			if argNode != nil {
				data[i] = *argNode
				newClaims = append(newClaims, claimNode)
				if node.ID == "L0Wv33MFQiuWVbWEKcELsA" {
					fmt.Println("----------------------------For L0Wv33MFQiuWVbWEKcELsA added claimNodeto newClaims")
					fmt.Println("----------------------------For L0Wv33MFQiuWVbWEKcELsA created argNode", argNode.ID)
				}

				argument := NewArgument(*argNode)
				argument.ClaimID = claim.ID
				args[argument.ID] = argument
				createItem(colArgs, argument)
			}
		}
	}
	data = append(data, newClaims...)

	// Open collections for edges
	edgeInferences := openCollection(db, "inferences", true)
	edgeBaseClaims := openCollection(db, "base_claims", true)
	edgePremises := openCollection(db, "premises", true)

	// Second pass: create edges
	for _, node := range data {
		fmt.Printf("Read item for edges: %+v\n", node)
		switch node.Type {
		case NODE_TYPE_CLAIM:
			nodeClaim, ok := claims[node.ID]
			if !ok {
				panic(fmt.Sprintf("Node claim %s not found", node.ID))
			}
			if node.MultiPremise {
				if len(node.Children) == 0 {
					fmt.Println("----------------------------MPClaim has no children")
				}
				for key, childVal := range node.Children {
					child := NewChildFromData(key, childVal)
					if child != nil {
						if claim, ok := claims[child.ID]; ok {
							createPremise(edgePremises, nodeClaim.ArangoID(), claim, node.ChildOrder(child.ID))
						} else {
							panic(fmt.Sprintf("Child Premise %s not found", child.ID))
						}
					} else {
						fmt.Println("----------------------------Premise child from data is nil")
					}
				}
			} else {
				if len(node.Children) == 0 {
					fmt.Println("----------------------------Claim has no children")
				}
				for key, childVal := range node.Children {
					child := NewChildFromData(key, childVal)
					if child != nil {
						id := nodeClaim.ID
						if arg, ok := args[child.ID]; ok {
							arg.TargetClaimID = &id
							arg.Pro = child.IsPro()
							updateItem(colArgs, arg.Key, arg)
							args[child.ID] = arg
							createInference(edgeInferences, nodeClaim.ArangoID(), arg)
						} else if claim, ok := claims[child.ID]; ok {
							// Data consistency problem in the Debate Map version!
							// Create an intervening Argument to resolve the problem
							arg := Argument{
								Key:           uuid.New().String(),
								ID:            child.ID,
								TargetClaimID: &id,
								ClaimID:       claim.ID,
								CreatedAt:     claim.CreatedAt,
								Creator:       claim.Creator,
								Pro:           child.IsPro(),
								Relevance:     1.00,
								Str:           0.50,
							}
							createItem(colArgs, arg)
							createInference(edgeInferences, nodeClaim.ArangoID(), arg)
							createBaseClaim(edgeBaseClaims, arg, claim.ArangoID())
						} else {
							panic(fmt.Sprintf("Child Argument %s not found", child.ID))
						}
					} else {
						fmt.Println("----------------------------Claim child from data is nil")
					}
				}
			}
		case NODE_TYPE_ARGUMENT:
			nodeArg, ok := args[node.ID]
			if !ok {
				panic(fmt.Sprintf("Node argument %s not found", node.ID))
			}
			if len(node.Children) == 0 {
				fmt.Println("----------------------------Argument has no children")
			}
			for key, childVal := range node.Children {
				child := NewChildFromData(key, childVal)
				if child != nil {
					id := nodeArg.ID
					if arg, ok := args[child.ID]; ok {
						arg.TargetArgumentID = &id
						arg.Pro = child.IsPro()
						updateItem(colArgs, arg.Key, arg)
						args[child.ID] = arg
						createInference(edgeInferences, nodeArg.ArangoID(), arg)
					} else if claim, ok := claims[child.ID]; ok {
						nodeArg.ClaimID = claim.ID
						updateItem(colArgs, nodeArg.Key, nodeArg)
						args[node.ID] = nodeArg
						createBaseClaim(edgeBaseClaims, nodeArg, claim.ArangoID())
					} else {
						panic(fmt.Sprintf("Child %s not found", child.ID))
					}
				} else {
					fmt.Println("----------------------------Argument child from data is nil")
				}
			}
		}
	}

	fmt.Println("Done.")
}

func OpenArangoConnection(server, dbname, username, password string) (driver.Database, error) {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{server},
	})
	fmt.Println("Connecting to the database:", server)
	if err != nil {
		fmt.Println("Error connecting the the database:", err.Error())
		panic(err.Error())
	}
	conn, err = conn.SetAuthentication(driver.BasicAuthentication(username, password))
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

	fmt.Println("Choosing the database:", dbname)
	db, err := c.Database(nil, dbname)
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

func updateItem(c driver.Collection, key string, item interface{}) {
	meta, err := c.UpdateDocument(nil, key, item)
	if err != nil {
		fmt.Printf("Error updating item: %s\nItem: %+v\n", err.Error(), item)
		panic(err.Error())
	}
	fmt.Println("Updated item. Meta:", meta)
}

func createInference(c driver.Collection, fromid string, toArg Argument) {
	inference := Inference{
		Key:       uuid.New().String(),
		CreatedAt: toArg.CreatedAt,
		Creator:   toArg.Creator,
		From:      fromid,
		To:        toArg.ArangoID(),
	}
	createItem(c, inference)
}

func createPremise(c driver.Collection, fromid string, toClaim Claim, order int) {
	premise := Premise{
		Key:       uuid.New().String(),
		CreatedAt: toClaim.CreatedAt,
		Creator:   toClaim.Creator,
		From:      fromid,
		To:        toClaim.ArangoID(),
		Order:     order,
	}
	createItem(c, premise)
}

func createBaseClaim(c driver.Collection, fromArg Argument, toid string) {
	bc := BaseClaim{
		Key:       uuid.New().String(),
		CreatedAt: fromArg.CreatedAt,
		Creator:   fromArg.Creator,
		From:      fromArg.ArangoID(),
		To:        toid,
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
