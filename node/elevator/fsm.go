package elevator

func F_fsmFloorArrival(newFloor int8, elevator T_Elevator) T_Elevator{
	elevator.P_info.Floor = newFloor
	SetFloorIndicator(int(newFloor))

	switch elevator.P_info.State {
	case MOVING:
		if F_shouldStop(elevator) {
			elevator = F_clearRequest(elevator)
		}
	case IDLE:
		SetMotorDirection(MD_Stop)
	}
	return elevator
}

func F_ReceiveRequest(req T_Request, elevator T_Elevator) T_Elevator{
	switch elevator.P_info.State {
	case IDLE:
		elevator.P_serveRequest = &req
		//spør heimvik om request sin state skal endres til active
		elevator = F_chooseDirection(elevator)
	}
	return elevator
}

func F_sendRequest(button ButtonEvent, requestOut chan T_Request, elevator T_Elevator) {
	elevator.CurrentID++
	if button.Button == BT_Cab {
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: CAB, Floor: int8(button.Floor)}
		return
	} else {
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: HALL, Floor: int8(button.Floor)}
		return
	}
}
