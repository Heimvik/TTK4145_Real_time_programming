package elevator

func F_FloorArrival(newFloor int8, elevator T_Elevator) T_Elevator {
	elevator.P_info.Floor = newFloor
	switch elevator.P_info.State {
	case MOVING:
		if F_shouldStop(elevator) {
			elevator = F_SetElevatorDirection(elevator) //COMMENT: Ta inn request her?
		}
	case IDLE: //should only happen when initializing, when the elevator first reaches a floor
		F_SetMotorDirection(NONE)
	}
	return elevator
}

func F_DoorTimeout(elevator T_Elevator, c_requestOut chan T_Request) (T_Elevator, T_Request) { //COMMENT: c_requestOut ute i run_elevator (her brukes den ikke da)
	if elevator.P_info.State == DOOROPEN && !elevator.Obstructed { //hvis heisen ikke er obstructed skal den gå til IDLE
		elevator.P_info.State = IDLE
		return elevator, T_Request{}
	} else if (elevator.P_info.State == DOOROPEN) && (elevator.Obstructed) && (elevator.P_serveRequest != nil) {
		resendReq := *elevator.P_serveRequest
		resendReq.State = UNASSIGNED
		return elevator, resendReq
	}
	return elevator, T_Request{} 
}

// her mottar jeg (kan oppstå deadlock)
// COMMENT: Samle i sentral FSM
func F_ReceiveRequest(req T_Request, elevator T_Elevator) T_Elevator {
	elevator.P_serveRequest = &req
	elevator.P_serveRequest.State = ACTIVE
	elevator = F_SetElevatorDirection(elevator)
	return elevator
}


// COMMENT: Legg ut i Run_elevator
func F_SendRequest(button T_ButtonEvent, requestOut chan T_Request, elevator T_Elevator) {
func F_SendRequest(button T_ButtonEvent, requestOut chan T_Request, elevator T_Elevator) {
	if button.Button == BT_Cab {
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: CAB, Floor: int8(button.Floor)}
	} else {
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: HALL, Floor: int8(button.Floor)}
	}
}
