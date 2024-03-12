package elevator

import (
	"fmt"
)

func F_FSM(c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	for {
		select {
		case button := <-chans.C_buttons:
			fmt.Print("Beginning button event\n")
			f_HandleButtonEvent(button, c_getSetElevatorInterface, chans)
			fmt.Print("Finished button event\n")
		case newFloor := <-chans.C_floors:
			fmt.Print("Beginning floor event\n")
			f_HandleFloorArrivalEvent(int8(newFloor), c_getSetElevatorInterface, chans)
			fmt.Print("Finished floor event\n")
		case <-chans.C_timerTimeout:
			fmt.Print("Beginning door timeout event\n")
			f_HandleDoorTimeoutEvent(c_getSetElevatorInterface, chans)
			fmt.Print("Finished door timeout event\n")
		case newRequest := <-chans.C_requestIn:
			fmt.Print("Beginning request to elevator event\n")
			f_HandleRequestToElevatorEvent(newRequest, c_getSetElevatorInterface, chans)
			fmt.Print("Finished request to elevator event\n")
		case obstructed := <-chans.C_obstr:
			f_HandleObstructedEvent(obstructed, c_getSetElevatorInterface, chans)
		case stop := <-chans.C_stop:
			f_HandleStopEvent(stop, c_getSetElevatorInterface, chans)
		}
	}
}

func f_HandleButtonEvent(button T_ButtonEvent, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	oldElevator.CurrentID++
	chans.getSetElevatorInterface.C_set <- oldElevator
	F_SendRequest(button, chans.C_requestOut, oldElevator)
}

//riktig floor -> stopelevator -> clear request ->  send done -> open door
//feil floor -> setDir
//IDLE||DOOROPEN -> stopelevator
func f_HandleFloorArrivalEvent(newFloor int8, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	F_SetFloorIndicator(int(newFloor))
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := F_FloorArrival(newFloor, oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator
	//JONASCOMMENT: sjekk om logikken her kan forenkles
	if newElevator.P_info.State == DOOROPEN { //legg inn mer direkte, som ikke er avhengig av det forrige her?
		F_SetDoorOpenLamp(true)
		fmt.Print("Sending done to node\n")
		oldElevator.P_serveRequest.State = DONE
		chans.C_requestOut <- *oldElevator.P_serveRequest
		chans.C_timerStart <- true
	}
}

func f_HandleDoorTimeoutEvent(c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := F_DoorTimeout(oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator
	if newElevator.P_info.State == IDLE {
		chans.C_timerStop <- true
		F_SetDoorOpenLamp(false)
	} else {
		chans.C_timerStart <- true
	}
}

//samme floor -> clearReq -> send ACTIVE -> send DONE -> open door
//forskjellig floor -> setDir -> send active
func f_HandleRequestToElevatorEvent(newRequest T_Request, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := F_ReceiveRequest(newRequest, oldElevator)
	cleared := false
	if F_ShouldStop(newElevator){
		newElevator = F_ClearRequest(newElevator)
		cleared = true
	} else {
		newElevator = F_SetElevatorDirection(newElevator)
	}
	chans.getSetElevatorInterface.C_set <- newElevator

	if cleared {
		fmt.Print("Sending active to node\n")
		newRequest.State = ACTIVE
		chans.C_requestOut <- newRequest
		fmt.Print("Sending done to node\n")
		newRequest.State = DONE
		chans.C_requestOut <- newRequest
		F_SetDoorOpenLamp(true)
		chans.C_timerStart <- true
	} else {
		fmt.Print("Sending active to node\n")
		newRequest.State = ACTIVE
		chans.C_requestOut <- newRequest
	}
}


func f_HandleObstructedEvent(obstructed bool, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	oldElevator.Obstructed = obstructed
	chans.getSetElevatorInterface.C_set <- oldElevator
}

func f_HandleStopEvent(stop bool, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get

	oldElevator.StopButton = stop
	F_SetStopLamp(stop)
	if stop {
		oldElevator = F_StopElevator(oldElevator)
	} else {
		oldElevator = F_SetElevatorDirection(oldElevator)
	}
	chans.getSetElevatorInterface.C_set <- oldElevator
}

func F_FloorArrival(newFloor int8, elevator T_Elevator) T_Elevator {
	elevator.P_info.Floor = newFloor
	switch elevator.P_info.State {
	case MOVING:
		if F_ShouldStop(elevator) {
			elevator = F_StopElevator(elevator)
			elevator = F_ClearRequest(elevator)
		}
	// case IDLE: //should only happen when initializing, when the elevator first reaches a floor
	default: //changed to default, in case of elevator being moved during dooropen
		elevator = F_StopElevator(elevator)
	}
	return elevator
}

func F_DoorTimeout(elevator T_Elevator) T_Elevator {
	if elevator.P_info.State == DOOROPEN && !elevator.Obstructed {
		elevator.P_info.State = IDLE
	}
	return elevator
}