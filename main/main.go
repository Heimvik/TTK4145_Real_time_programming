package main

import (
	"fmt"
	"the-elevator/node"
)

// Main file of a node
func main() {
	fmt.Printf("Hallo world\n")
	var testNode node.Node //Note the syntax for access to other packages
	testNode.Priority = 1
	fmt.Println(testNode.Priority)
}
