package node

import (
	"the-elevator/elevator"
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
func f_AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func f_FindClosestElevator(floor int, elevators []*elevator.T_Elevator) *elevator.T_Elevator {
	var closestElevator *elevator.T_Elevator
	closestFloor := FLOORS
	for _, p_elevator := range elevators {
		currentDifference := f_AbsInt(p_elevator.Floor - floor)
		if currentDifference < closestFloor {
			closestFloor = p_elevator.Floor
			closestElevator = p_elevator
		}
	}
	return closestElevator
}

func F_AssignRequest(request *elevator.T_Request, connectedElevators []*elevator.T_Elevator) *elevator.T_Elevator {
	//Hall and Cab requests are assigned differently:
	//Hall: assigned to a new connected elevator no matter what
	//Cab: redistribution in event of a door sensor, are on the layer outside this function
	//The only this function should do is to assign the request to the best avalebale elevators in connectedElevators
	var avalibaleElevators []*elevator.T_Elevator
	for _, p_elevator := range connectedElevators {
		if p_elevator.State == elevator.EB_Idle {
			avalibaleElevators = append(avalibaleElevators, p_elevator)
		}
	}

	var returnElevator *elevator.T_Elevator
	switch request.Calltype {
	case elevator.Hall:
		returnElevator = f_FindClosestElevator(request.Floor, avalibaleElevators)
	case elevator.Cab:
		if request.P_Elevator.State == elevator.EB_Idle {
			returnElevator = request.P_Elevator
		} else {
			returnElevator = f_FindClosestElevator(request.Floor, avalibaleElevators)
		}
	}
	return returnElevator
}
