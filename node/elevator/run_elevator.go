package elevator

import (
	"fmt"
)

/*
Runs the elevator, initializing channels, polling sensors, managing timers, and running the finite state machine (FSM).

Prerequisites: The elevatorOperations parameter must be initialized with valid elevator operations. The elevatorport parameter must be a valid port number. The c_elevatorWithoutErrors channel must be initialized for error reporting.

Returns: Nothing, but continuously runs the elevator.
*/

func F_RunElevator(elevatorOperations T_ElevatorOperations, c_getSetElevatorInterface chan T_GetSetElevatorInterface, c_requestOut chan T_Request, c_requestIn chan T_Request, elevatorport int, c_elevatorWithoutErrors chan bool) {
	F_InitDriver(fmt.Sprintf("localhost:%d", elevatorport))

	//force elevator to move in case of starting between floors
	F_SetMotorDirection(ELEVATORDIRECTION_DOWN)

	var chans T_ElevatorChannels = F_InitChannels(c_requestIn, c_requestOut)
	//interface for getting and setting elevator
	go F_GetAndSetElevator(elevatorOperations, c_getSetElevatorInterface)
	//polling sensors
	go F_PollButtons(chans.C_buttons)
	go F_PollFloorSensor(chans.C_floors)
	go F_PollObstructionSwitch(chans.C_obstr)
	go F_PollStopButton(chans.C_stop)
	//doortimer
	go F_DoorTimer(chans)
	go F_EnsureElevatorDirection(chans, c_getSetElevatorInterface)
	//FSM
	go F_FSM(c_getSetElevatorInterface, chans, c_elevatorWithoutErrors)
}
