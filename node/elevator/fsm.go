package elevator

func F_fsmFloorArrival(newFloor int) {
	Elevator.P_info.Floor = newFloor
	SetFloorIndicator(newFloor)

	switch Elevator.P_info.State {
	case MOVING:
		if F_shouldStop(Elevator) {
			//make timer logic so door stays open for as long as it should
			F_clearRequest(Elevator)
		}
	// case DOOROPEN:
	// 	//make timer logic so door stays open for as long as it should
	// 	time.Sleep(3 * time.Second) //placeholder
	// 	Elevator.P_info.State = IDLE
	// 	F_clearRequest(Elevator)

	case IDLE:
		SetMotorDirection(MD_Stop)
	}
}

func F_fsmObstructionSwitch(obstructed bool) {
	switch Elevator.P_info.State {
	case DOOROPEN:
		if obstructed == false {
			Elevator.P_info.State = IDLE
		}
	}
}

func F_fsmDoorTimeout() {
	switch Elevator.P_info.State {
	case DOOROPEN:
		if C_obstruction == false {
			Elevator.P_info.State = IDLE
		}
	}
}

func F_ReceiveRequest(req T_Request) {
	switch Elevator.P_info.State {
	case IDLE:
		Elevator.P_serveRequest = &req
		F_chooseDirection(Elevator)
	}
}

func F_sendRequest(button ButtonEvent, requestOut chan T_Request) {
	Elevator.currentID++
	if button.Button == BT_Cab {
		requestOut <- T_Request{Id: Elevator.currentID, State: 0, Calltype: CAB, Floor: button.Floor}
		return
	} else {
		requestOut <- T_Request{Id: Elevator.currentID, State: 0, Calltype: HALL, Floor: button.Floor}
		return
	}
}