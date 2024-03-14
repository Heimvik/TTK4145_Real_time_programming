package elevator

import (
	"time"
)

/*
Initializes elevator channels.

Prerequisites: None

Returns: Initialized elevator channels.
*/
func F_InitChannels(c_requestIn chan T_Request, c_requestOut chan T_Request) T_ElevatorChannels {
	return T_ElevatorChannels{
		getSetElevatorInterface: T_GetSetElevatorInterface{C_get: make(chan T_Elevator), C_set: make(chan T_Elevator)},
		C_timerStart:            make(chan bool),
		C_timerStop:             make(chan bool),
		C_timerTimeout:          make(chan bool),
		C_buttons:               make(chan T_ButtonEvent),
		C_floors:                make(chan int),
		C_obstr:                 make(chan bool),
		C_stop:                  make(chan bool),
		C_requestIn:             c_requestIn,
		C_requestOut:            c_requestOut,
	}
}

/*
Retrieves the current elevator.

Prerequisites: The elevatorOperations parameter must be initialized with valid elevator operations.

Returns: The current elevator.
*/
func F_GetElevator(elevatorOperations T_ElevatorOperations) T_Elevator {
	c_responseChan := make(chan T_Elevator)
	elevatorOperations.C_getElevator <- c_responseChan // Send the response channel to the NodeOperationManager
	elevator := <-c_responseChan                       // Receive the connected nodes from the response channel
	return elevator
}

/*
Sets the elevator.

Prerequisites: None

Returns: Nothing
*/
func F_SetElevator(elevatorOperations T_ElevatorOperations, elevator T_Elevator) {
	elevatorOperations.C_setElevator <- elevator // Send the connectedNodes directly to be written
}

/*
Retrieves and sets the elevator state concurrently.

Prerequisites: The elevatorOperations parameter must be initialized with valid elevator operations.

Returns: Nothing
*/
func F_GetAndSetElevator(elevatorOperations T_ElevatorOperations, c_getSetElevatorInterface chan T_GetSetElevatorInterface) { //let run in a sepreate goroutine
	for {
	WAITFORINTERFACE:
		select {
		case elevatorInterface := <-c_getSetElevatorInterface:
			c_responsChan := make(chan T_Elevator)
			elevatorOperations.C_getSetElevator <- c_responsChan
			getSetTimer := time.NewTicker(time.Duration(5) * time.Second)
			for {
				select {
				case oldElevator := <-c_responsChan:
					elevatorInterface.C_get <- oldElevator
				case newElevator := <-elevatorInterface.C_set:
					c_responsChan <- newElevator
					break WAITFORINTERFACE
				case <-getSetTimer.C:
					break WAITFORINTERFACE
				}
			}
		}
	}
}

/*
Determines if the elevator should stop at the current floor.

Prerequisites: Elevator needs to currently have a request to serve.

Returns: True if the elevator should stop; otherwise, false.
*/
func F_ShouldElevatorStop(elevator T_Elevator) bool {
	return (elevator.P_info.Floor == elevator.ServeRequest.Floor)
}

/*
Stops the elevator.

Prerequisites: None

Returns: The stopped elevator.
*/
func F_StopElevator(elevator T_Elevator) T_Elevator {
	elevator.P_info.Direction = ELEVATORDIRECTION_NONE
	F_SetMotorDirection(ELEVATORDIRECTION_NONE)
	return elevator
}

/*
Chooses the direction for the elevator to move based on the requested floor.

Prerequisites: None

Returns: The updated elevator state.
*/
func F_ChooseElevatorDirection(elevator T_Elevator) T_Elevator {
	if elevator.ServeRequest.Floor < elevator.P_info.Floor {
		elevator.P_info.State = ELEVATORSTATE_MOVING
		elevator.P_info.Direction = ELEVATORDIRECTION_DOWN
		F_SetMotorDirection(ELEVATORDIRECTION_DOWN)
	} else if elevator.ServeRequest.Floor > elevator.P_info.Floor {
		elevator.P_info.State = ELEVATORSTATE_MOVING
		elevator.P_info.Direction = ELEVATORDIRECTION_UP
		F_SetMotorDirection(ELEVATORDIRECTION_UP)
	} else {
		elevator.P_info.Direction = ELEVATORDIRECTION_NONE
		F_SetMotorDirection(ELEVATORDIRECTION_NONE)
	}
	return elevator
}

/*
Ensures the elevator is moving in the correct direction by continuously polling the elevator state and setting motordirection accordingly.

Prerequisites: None

Returns: Nothing, but continuously sets the motor direction based on the elevator state.
*/
func F_EnsureElevatorDirection(chans T_ElevatorChannels, c_getSetElevatorInterface chan T_GetSetElevatorInterface) {
	for{
		c_getSetElevatorInterface <- chans.getSetElevatorInterface
		elevator := <-chans.getSetElevatorInterface.C_get
		chans.getSetElevatorInterface.C_set <- elevator

		if elevator.P_info.Direction == ELEVATORDIRECTION_DOWN {
			F_SetMotorDirection(ELEVATORDIRECTION_DOWN)
		} else if elevator.P_info.Direction == ELEVATORDIRECTION_UP {
			F_SetMotorDirection(ELEVATORDIRECTION_UP)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
