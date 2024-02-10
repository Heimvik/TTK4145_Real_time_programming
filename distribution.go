package node

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

//Should give out the elevator that should serve the request

func AssignRequest(var request Request, var connectedElevators[] Elevator){
	//Hall and Cab requests are assigned differently:
	//Hall: assigned to a new connected elevator no matter what
	//Cab: redistribution in event of a door sensor, are on the layer outside this function
	//The only this function should do is to assign the request to the avalebale elevaots in connectedElevaots
	var avalibaleElevators[] Elevator;
	for elevator:=0;len(connectedElevators);elevator++{
		if(elevator.avalibale){
			avalibaleElevators
		}
	}
	case
}



