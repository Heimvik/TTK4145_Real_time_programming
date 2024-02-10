package node

import (
	"fmt"
)

//Should take in:
//The request struct containing:
// - cab/hall
// - elevator id
// - floor
// - diraction
//
//Connected elevators and their struct:
// - elevator id
// - state (idle, moving, dooropen)
// - floor
// - diraction

// Should give out the elevator that should serve the request
func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func AssignRequest(request Request, connectedElevators []Elevator) {
	//Hall and Cab requests are assigned differently:
	//Hall: assigned to a new connected elevator no matter what
	//Cab: redistribution in event of a door sensor, are on the layer outside this function
	//The only this function should do is to assign the request to the best avalebale elevators in connectedElevators
	var avalibaleElevators []Elevator
	for _, elevator := range connectedElevators {
		if elevator.Avalibale {
			avalibaleElevators = append(avalibaleElevators, elevator)
		}
	}

	switch request.Calltype {
	case hall:
		closestFloor := FLOORS
		for _, elevator := range avalibaleElevators {
			currentDifference := AbsInt(elevator.Floor - request.Floor)
			if currentDifference < closestFloor {
				closestFloor := elevator.Floor
				fmt.Println(closestFloor)
			} else if currentDifference == closestFloor {
				fmt.Println(currentDifference)
			}
		}

	case cab:
	}
}
