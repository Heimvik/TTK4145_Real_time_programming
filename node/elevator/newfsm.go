package elevator



func F_FSM(c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	for{
		select {
		case button := <-chans.C_buttons:
			f_HandleButtonEvent(button, c_getSetElevatorInterface, chans)
		case newFloor := <-chans.C_floors:
			f_HandleFloorArrivalEvent(int8(newFloor), c_getSetElevatorInterface, chans)
		case <-chans.C_timerTimeout:
			f_HandleDoorTimeoutEvent(c_getSetElevatorInterface, chans)
		case newRequest := <-chans.C_requestIn:
			f_HandleRequestToElevatorEvent(newRequest, c_getSetElevatorInterface, chans)
		case obstructed := <-chans.C_obstr:
			f_HandleObstructedEvent(obstructed, c_getSetElevatorInterface, chans)
		case stop := <-chans.C_stop:
			f_HandleStopEvent(stop, c_getSetElevatorInterface, chans)
		}
	}
}

//function for handling buttonEvent
func f_HandleButtonEvent(button T_ButtonEvent, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels){
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	oldElevator.CurrentID++
	chans.getSetElevatorInterface.C_set <- oldElevator
	F_SendRequest(button, chans.C_requestOut, oldElevator)
}

//function for handling floorArrivalEvent
func f_HandleFloorArrivalEvent(newFloor int8, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels){
	
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := F_FloorArrival(newFloor, oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator
	//JONASCOMMENT: sjekk om logikken her kan forenkles
	if newElevator.P_info.State == DOOROPEN {
		go F_DoorTimer(chans.C_timerStop, chans.C_timerTimeout)
	}
	if newElevator.P_info.Direction == NONE && !newElevator.StopButton && oldElevator.P_serveRequest != nil {
		oldElevator.P_serveRequest.State = DONE
		chans.C_requestOut <- *oldElevator.P_serveRequest
	}
}

//function for handling doorTimeoutEvent
func f_HandleDoorTimeoutEvent(c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels){
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator, newReq := F_DoorTimeout(oldElevator, chans.C_requestOut)
	chans.getSetElevatorInterface.C_set <- newElevator
	//JONASCOMMENT: sjekk om logikken her kan forenkles
	if newReq.State == UNASSIGNED && newElevator.P_serveRequest != nil {
		chans.C_requestOut <- newReq
	} else if newElevator.P_info.State == IDLE {
		chans.C_timerStop <- true
	}
}

//function for handling requestToElevatorEvent
func f_HandleRequestToElevatorEvent(newRequest T_Request, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels){
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := F_ReceiveRequest(newRequest, oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator

	if newElevator.P_info.State == DOOROPEN {
		newRequest.State = ACTIVE
		chans.C_requestOut <- newRequest
		newRequest.State = DONE
		chans.C_requestOut <- newRequest
		go F_DoorTimer(chans.C_timerStop, chans.C_timerTimeout)
	} else {
		chans.C_requestOut <- *newElevator.P_serveRequest
	}
}

//function for handling obstructedEvent
func f_HandleObstructedEvent(obstructed bool, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels){
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	oldElevator.Obstructed = obstructed
	chans.getSetElevatorInterface.C_set <- oldElevator
}

//function for handling stopEvent
func f_HandleStopEvent(stop bool, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels){
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	oldElevator.StopButton = stop
	newElevator := F_SetElevatorDirection(oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator
}