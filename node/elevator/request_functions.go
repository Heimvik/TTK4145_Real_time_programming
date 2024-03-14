package elevator

/*
Converts a request into a button type.

Prerequisites: None

Returns: The converted button type.
*/
func F_ConvertRequestToButtonType(request T_Request) T_ButtonType {
	if request.Calltype == CALLTYPE_HALL {
		if request.Direction == ELEVATORDIRECTION_UP {
			return BUTTONTYPE_HALLUP
		} else if request.Direction == ELEVATORDIRECTION_DOWN {
			return BUTTONTYPE_HALLDOWN
		}
	} else if request.Calltype == CALLTYPE_CAB {
		return BUTTONTYPE_CAB
	}
	return 0
}

/*
Converts a button event into a request.

Prerequisites: None

Returns: The converted request.
*/
func F_ConvertButtonTypeToRequest(buttonEvent T_ButtonEvent) T_Request {
	if buttonEvent.Button == BUTTONTYPE_HALLUP {
		return T_Request{Floor: int8(buttonEvent.Floor), Calltype: CALLTYPE_HALL, Direction: ELEVATORDIRECTION_UP}
	} else if buttonEvent.Button == BUTTONTYPE_HALLDOWN {
		return T_Request{Floor: int8(buttonEvent.Floor), Calltype: CALLTYPE_HALL, Direction: ELEVATORDIRECTION_DOWN}
	} else if buttonEvent.Button == BUTTONTYPE_CAB {
		return T_Request{Floor: int8(buttonEvent.Floor), Calltype: CALLTYPE_CAB, Direction: ELEVATORDIRECTION_NONE}
	}
	return T_Request{}
}

/*
Receives a request and updates the elevator state.

Prerequisites: None

Returns: The updated elevator state.
*/
func F_ReceiveRequest(req T_Request, elevator T_Elevator) T_Elevator {
	elevator.ServeRequest = req
	elevator.ServeRequest.State = REQUESTSTATE_ACTIVE
	return elevator
}

/*
Clears the current request from the elevator.

Prerequisites: None

Returns: The updated elevator state.
*/
func F_ClearRequest(elevator T_Elevator) T_Elevator {
	elevator.ServeRequest = T_Request{}
	elevator.P_info.State = ELEVATORSTATE_DOOROPEN
	return elevator
}

/*
Sends a request to the node based on the button event.

Prerequisites: None

Returns: Nothing, but sends the request to the requestOut channel.
*/
func F_SendRequestToNode(button T_ButtonEvent, requestOut chan T_Request, elevator T_Elevator) {
	request := F_ConvertButtonTypeToRequest(button)
	if button.Button == BUTTONTYPE_CAB {
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: request.Calltype, Floor: request.Floor, Direction: request.Direction}
	} else {
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: request.Calltype, Floor: request.Floor, Direction: request.Direction}
	}
}
