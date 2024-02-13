package main

import (
	"the-elevator/node"
)

// Main file of a node
func main() {
	node.F_Init("default")
	for{
		node.F_Run()
	}
	
}

//Spørsmål til studasser:
//Nødvendig med FSM for sende/motta
//Acceptance funksjoner for å motta VIKTIG
//Fordeling av globale typer
