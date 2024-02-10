package main

import (
	"fmt"
	"the-elevator/node"
)

// Main file of a node
func main() {
	var elevators []*node.T_Elevator
	for i := 0; i <= 2; i++ {
		newElevator := node.T_Elevator{
			Floor:     4 - i,
			Direction: node.Down,
			Avalibale: true,
		}
		if i == 2 {
			newElevator.Avalibale = false
		}
		elevators = append(elevators, &newElevator)
	}

	request := node.T_Request{
		Calltype:   node.Hall,
		P_Elevator: elevators[0],
		Floor:      1, //elevators[0].Floor,
		Direction:  node.Up,
	}

	fmt.Println(node.F_AssignRequest(&request, elevators).Floor)

}
