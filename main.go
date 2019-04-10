package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func main() {
	filename := "data/Test1.json"
	fmt.Println("Loading file:", filename)

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error loading file:", err.Error())
	}

	data := []DebateMapNode{}
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		fmt.Println("Error parsing JSON:", err.Error())
	}

	for _, item := range data {
		fmt.Printf("Read item: +%v\n", item)
	}

	fmt.Println("Done.")
}
