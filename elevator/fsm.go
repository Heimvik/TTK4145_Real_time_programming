package elevator

func F_fsmFloorArrival(newFloor int) {

	Elevator.Floor = newFloor
	// SetFloorIndicator(newFloor)

	switch Elevator.State {
	case EB_Moving:
		if F_shouldStop(Elevator){
			SetMotorDirection(MD_Stop)
			Elevator.State = EB_DoorOpen
			//make timer logic so door stays open for as long as it should
			F_clearRequest(Elevator)
		}

	case EB_DoorOpen:
		SetMotorDirection(MD_Stop)
		//again, start a timer
		F_clearRequest(Elevator)
	}

}

//under will be func when Door is closing, this will either make the elevator start moving again if there is a new request, or set the elevators state to idle
//func F_fsmDoorClosing()

func F_fsmButtonPress(ButtonEvent) {
	

}
