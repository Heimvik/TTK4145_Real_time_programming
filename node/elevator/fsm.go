package elevator

func F_fsmFloorArrival(newFloor int8, elevator T_Elevator, c_requestOut chan T_Request) T_Elevator{
	elevator.P_info.Floor = newFloor
	switch elevator.P_info.State {
	case MOVING:
		if F_shouldStop(elevator) {
			elevator = F_clearRequest(elevator, c_requestOut)
		}
	case IDLE: //should only happen when initializing, when the elevator first reaches a floor
		SetMotorDirection(MD_Stop)
	}
	return elevator
}

func F_fsmDoorTimeout(elevator T_Elevator, c_requestOut chan T_Request) T_Elevator{
	if elevator.P_info.State == DOOROPEN && !elevator.Obstructed { //hvis heisen ikke er obstructed skal den gå til IDLE 
		elevator.P_info.State = IDLE
	} else if (elevator.P_info.State == DOOROPEN) && (elevator.Obstructed) && (elevator.P_serveRequest != nil) { //hvis heisen er obstructed skal den fortsette å være DOOROPEN
		resendReq := *elevator.P_serveRequest
		resendReq.State = UNASSIGNED
		c_requestOut <- resendReq
	}
	return elevator
}

func F_ReceiveRequest(req T_Request, elevator T_Elevator, c_requestOut chan T_Request) T_Elevator{
	switch elevator.P_info.State {
	case IDLE:
		elevator.P_serveRequest = &req
		elevator.P_serveRequest.State = ACTIVE
		elevator = F_chooseDirection(elevator, c_requestOut)
		if elevator.P_info.State == MOVING {
			c_requestOut <- *elevator.P_serveRequest
		}
	}
	return elevator
}

func F_sendRequest(button ButtonEvent, requestOut chan T_Request, elevator T_Elevator) T_Elevator {
	elevator.CurrentID++
	if button.Button == BT_Cab {
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: CAB, Floor: int8(button.Floor)}
		return elevator
	} else {
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: HALL, Floor: int8(button.Floor)}
		return elevator
	}
}