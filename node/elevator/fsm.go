package elevator

import (
	"time"
)

/*
Runs the finite state machine (FSM), managing various elevator events such as button presses, floor arrivals, door timeouts, requests to the elevator, obstructions, and stop button events.

Prerequisites: The c_getSetElevatorInterface channel, T_ElevatorChannels struct, and c_elevatorWithoutErrors channel must be properly initialized.

Returns: Nothing, but continuously listens for elevator events and updates the elevator state accordingly.
*/
func F_FSM(c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels, c_elevatorWithoutErrors chan bool) {
	for {
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
		default:
			c_elevatorWithoutErrors <- true
			time.Sleep(time.Duration(1) * time.Millisecond)
		}
	}
}

/*
Generates a new request based on the button pressed, and increments elevator.CurrentID.

Prerequisites: Nothing

Returns: Nothing, but sends the updated elevator state to the c_getSetElevatorInterface channel, and sends a request to the Node module.
*/
func f_HandleButtonEvent(button T_ButtonEvent, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	oldElevator.CurrentID++
	chans.getSetElevatorInterface.C_set <- oldElevator

	F_SendRequestToNode(button, chans.C_requestOut, oldElevator)
}

/*
Updates the floor indicator and the elevators current floor to match the arriving floor, and checks if the elevator can clear a request at the current floor.

Prerequisites: Nothing

Returns: Nothing, but sends the updated elevator state to the c_getSetElevatorInterface channel, and sends a request to the Node module if the elevator clears a request at the current floor.
*/
func f_HandleFloorArrivalEvent(newFloor int8, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	F_SetFloorIndicator(int(newFloor))
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := f_UpdateElevatorOnFloorArrival(newFloor, oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator

	if newElevator.P_info.State == ELEVATORSTATE_DOOROPEN {
		F_SetDoorOpenLamp(true)
		if (oldElevator.ServeRequest != T_Request{}) {
			oldElevator.ServeRequest.State = REQUESTSTATE_DONE
			chans.C_requestOut <- oldElevator.ServeRequest
		}
		chans.C_timerStart <- true
	}
}

/*
Tries to close the elevator door, and stops the door timer if the elevator is idle.

Prerequisites: Nothing

Returns: Nothing, but sends the updated elevator state to the c_getSetElevatorInterface channel.
*/
func f_HandleDoorTimeoutEvent(c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := f_TryCloseDoor(oldElevator)
	chans.getSetElevatorInterface.C_set <- newElevator
	if newElevator.P_info.State == ELEVATORSTATE_IDLE {
		chans.C_timerStop <- true
		F_SetDoorOpenLamp(false)
	} else {
		chans.C_timerStart <- true
	}
}

/*
Updates the elevator state based on the request received from the Node module, and sends the updated state back to the Node module.

Prerequisites: Nothing

Returns: Nothing, but sends the updated elevator state to the c_getSetElevatorInterface channel, and sends a the requests current state back to the Node.
*/
func f_HandleRequestToElevatorEvent(newRequest T_Request, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	newElevator := F_ReceiveRequest(newRequest, oldElevator)
	cleared := false
	if F_ShouldElevatorStop(newElevator) {
		newElevator = F_ClearRequest(newElevator)
		cleared = true
	} else {
		newElevator = F_ChooseElevatorDirection(newElevator)
	}
	chans.getSetElevatorInterface.C_set <- newElevator

	if cleared {
		newRequest.State = REQUESTSTATE_ACTIVE
		chans.C_requestOut <- newRequest
		newRequest.State = REQUESTSTATE_DONE
		chans.C_requestOut <- newRequest
		F_SetDoorOpenLamp(true)
		chans.C_timerStart <- true
	} else {
		newRequest.State = REQUESTSTATE_ACTIVE
		chans.C_requestOut <- newRequest
	}
}

/*
Updates the elevators obstruction variable based on the obstruction switch.

Prerequisites: Nothing

Returns: Nothing, but sends the updated elevator to the c_getSetElevatorInterface channel.
*/
func f_HandleObstructedEvent(obstructed bool, c_getSetElevatorInterface chan T_GetSetElevatorInterface, chans T_ElevatorChannels) {
	c_getSetElevatorInterface <- chans.getSetElevatorInterface
	oldElevator := <-chans.getSetElevatorInterface.C_get
	oldElevator.P_info.Obstructed = obstructed
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
		oldElevator = F_ChooseElevatorDirection(oldElevator)
	}
	chans.getSetElevatorInterface.C_set <- oldElevator
}

/*
Updates the elevators floor variable, and stops the elevator if it should stop at the current floor.

Prerequisites: Nothing

Returns: The updated elevator.
*/
func f_UpdateElevatorOnFloorArrival(newFloor int8, elevator T_Elevator) T_Elevator {
	elevator.P_info.Floor = newFloor
	switch elevator.P_info.State {
	case ELEVATORSTATE_MOVING:
		if (elevator.ServeRequest != T_Request{}) {
			if F_ShouldElevatorStop(elevator) {
				elevator = F_StopElevator(elevator)
				elevator = F_ClearRequest(elevator)
			} else {
				elevator = F_ChooseElevatorDirection(elevator)
			}
		} else {
			elevator = F_StopElevator(elevator)
			elevator.P_info.State = ELEVATORSTATE_IDLE
		}
	default:
		elevator = F_StopElevator(elevator)
		
	}
	return elevator
}

/*
Tries to close the door of the elevator.

Prerequisites: Nothing

Returns: The updated elevator.
*/
func f_TryCloseDoor(elevator T_Elevator) T_Elevator {
	if elevator.P_info.State == ELEVATORSTATE_DOOROPEN && !elevator.P_info.Obstructed {
		elevator.P_info.State = ELEVATORSTATE_IDLE
	}
	return elevator
}
