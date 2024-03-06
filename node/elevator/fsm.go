package elevator

import "fmt"

func F_fsmFloorArrival(newFloor int8, elevator T_Elevator, c_requestOut chan T_Request) T_Elevator {
	elevator.P_info.Floor = newFloor
	switch elevator.P_info.State {
	case MOVING:
		if F_shouldStop(elevator) {
			elevator = F_chooseDirection(elevator, c_requestOut)
		}
	case IDLE: //should only happen when initializing, when the elevator first reaches a floor
		F_SetMotorDirection(NONE)
	}
	return elevator
}

func F_fsmDoorTimeout(elevator T_Elevator, c_requestOut chan T_Request) (T_Elevator, T_Request) {
	if elevator.P_info.State == DOOROPEN && !elevator.Obstructed { //hvis heisen ikke er obstructed skal den gå til IDLE
		elevator.P_info.State = IDLE
		return elevator, T_Request{}
	} else if (elevator.P_info.State == DOOROPEN) && (elevator.Obstructed) && (elevator.P_serveRequest != nil) { //hvis heisen er obstructed skal den fortsette å være DOOROPEN
		resendReq := *elevator.P_serveRequest
		resendReq.State = UNASSIGNED
		return elevator, resendReq
	}
	return elevator, T_Request{}
}

// her mottar jeg (kan oppstå deadlock)
func F_ReceiveRequest(req T_Request, elevator T_Elevator, c_requestOut chan T_Request) T_Elevator {
	switch elevator.P_info.State {
	case IDLE:
		elevator.P_serveRequest = &req
		elevator.P_serveRequest.State = ACTIVE
		elevator = F_chooseDirection(elevator, c_requestOut)
	}
	return elevator
}

// her sender jeg ut (bør ha fiksa deadlock)
func F_sendRequest(button T_ButtonEvent, requestOut chan T_Request, elevator T_Elevator) {
	if button.Button == BT_Cab {
		fmt.Println("Sending cab request to node at floor: ", button.Floor, " with state: ", 0)
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: CAB, Floor: int8(button.Floor)}
		fmt.Println("Cab request sent")
	} else {
		fmt.Println("Sending hall request to node at floor: ", button.Floor, " with state: ", 0)
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: HALL, Floor: int8(button.Floor)}
		fmt.Println("Hall request sent")
	}

}
