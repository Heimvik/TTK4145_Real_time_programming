package elevator

import (
	"fmt"
	"time"
)

func F_FSM(c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	for {
		select {
		case button := <-chans.C_buttons:
			fmt.Printf("Handling button event\n")
			f_HandleButtonEvent(button, c_getSetElevatorInterface, chans)
		case newFloor := <-chans.C_floors:
			fmt.Printf("Arrived at floor %v",newFloor)
			f_HandleFloorArrivalEvent(int8(newFloor), c_getSetElevatorInterface, chans)
		case <-chans.C_timerTimeout:
			f_HandleDoorTimeoutEvent(c_getSetElevatorInterface, chans)
		case newRequest := <-chans.C_requestIn:
			fmt.Print("Received request from node\n")
			f_HandleRequestToElevatorEvent(newRequest, c_getSetElevatorInterface, chans)
		case obstructed := <-chans.C_obstr:
			fmt.Print("Handling OBSTRUCTED event\n")
			f_HandleObstructedEvent(obstructed, c_getSetElevatorInterface, chans)
		case stop := <-chans.C_stop:
			fmt.Print("Handling STOP event\n")
			f_HandleStopEvent(stop, c_getSetElevatorInterface, chans)
		}
	}
}

func f_HandleButtonEvent(button T_ButtonEvent, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	oldElevator.CurrentID++
	chans.getSetElevatorInterface.C_set <- oldElevator
	fmt.Print("Sending request to node\n")
	F_SendRequest(button, chans.C_requestOut, oldElevator)
	fmt.Print("Request sent to node\n")
}

func f_HandleFloorArrivalEvent(newFloor int8, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := F_FloorArrival(newFloor, oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator
	//JONASCOMMENT: sjekk om logikken her kan forenkles
	if newElevator.P_info.State == DOOROPEN { //legg inn mer direkte, som ikke er avhengig av det forrige her?
		fmt.Print("FADO: Sending DONE request to node\n")
		chans.C_timerStart <- true
		oldElevator.P_serveRequest.State = DONE
		chans.C_requestOut <- *oldElevator.P_serveRequest
	}
}

func f_HandleDoorTimeoutEvent(c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := F_DoorTimeout(oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator
	if newElevator.P_info.State == IDLE {
		chans.C_timerStop <- true
		time.Sleep(time.Duration(DOOROPENTIME/2) * time.Millisecond) //closing door
	} else {
		chans.C_timerStart <- true
	}

	//JONASCOMMENT: sjekk om logikken her kan forenkles
	// if newReq.State == UNASSIGNED && newElevator.P_serveRequest != nil {
	// 	chans.C_requestOut <- newReq
	// } else if newElevator.P_info.State == IDLE {
	// 	chans.C_timerStop <- true
	// }
}

func f_HandleRequestToElevatorEvent(newRequest T_Request, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := F_ReceiveRequest(newRequest, oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator

	if newElevator.P_info.State == DOOROPEN {
		fmt.Print("DO: Sending ACTIVE request to node\n")
		newRequest.State = ACTIVE
		chans.C_requestOut <- newRequest
		fmt.Print("DO: SENDING DONE request to node\n")
		newRequest.State = DONE
		chans.C_requestOut <- newRequest
		chans.C_timerStart <- true
	} else {
		fmt.Print("MOVING: Sending ACTIVE request to node\n")
		chans.C_requestOut <- *newElevator.P_serveRequest
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

	newElevator := F_SetElevatorDirection(oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator
}

func F_FloorArrival(newFloor int8, elevator T_Elevator) T_Elevator {
	elevator.P_info.Floor = newFloor
	switch elevator.P_info.State {
	case MOVING:
		if F_shouldStop(elevator) {
			elevator = F_SetElevatorDirection(elevator)
		}
	// case IDLE: //should only happen when initializing, when the elevator first reaches a floor
	default: //changed to default, in case of elevator being moved during dooropen
		F_SetMotorDirection(NONE)
	}
	return elevator
}

func F_DoorTimeout(elevator T_Elevator) T_Elevator {
	if elevator.P_info.State == DOOROPEN && !elevator.Obstructed {
		elevator.P_info.State = IDLE
	}
	return elevator
}

//gammel innmat i DoorTimeout, fjernet at den resender fordi elevator bør ikke få inn request når DOOROPEN
// if elevator.P_info.State == DOOROPEN && !elevator.Obstructed { //hvis heisen ikke er obstructed skal den gå til IDLE
// 	elevator.P_info.State = IDLE
// 	return elevator, T_Request{}
// } else if (elevator.P_info.State == DOOROPEN) && (elevator.Obstructed) && (elevator.P_serveRequest != nil) {
// 	resendReq := *elevator.P_serveRequest
// 	resendReq.State = UNASSIGNED
// 	return elevator, resendReq
// }
