package elevator

import (
	"fmt"
	"time"
)

type T_ElevatorState uint8
type T_ElevatorDirection int8

const (
	IDLE     T_ElevatorState = 0
	DOOROPEN T_ElevatorState = 1
	MOVING   T_ElevatorState = 2
)

const (
	UP   T_ElevatorDirection = 1
	DOWN T_ElevatorDirection = -1
	NONE T_ElevatorDirection = 0
)

//Keeping this in case of future improvements regarding secondary requirements,
//see single-elevator/elevator.go for inspiration

type T_Elevator struct {
	CurrentID      int
	Obstructed     bool
	StopButton     bool
	P_info         *T_ElevatorInfo //MUST be pointer to info (points to info stored in ThisNode.NodeInfo.ElevatorInfo)
	P_serveRequest *T_Request      //Pointer to the current request you are serviceing
}

type T_ElevatorInfo struct {
	Direction T_ElevatorDirection
	Floor     int8 //ranges from 1-4
	State     T_ElevatorState
}

type T_GetSetElevatorInterface struct {
	C_get chan T_Elevator
	C_set chan T_Elevator
}

type T_ElevatorOperations struct {
	C_getElevator    chan chan T_Elevator
	C_setElevator    chan T_Elevator
	C_getSetElevator chan chan T_Elevator
}

func F_GetElevator(ops T_ElevatorOperations) T_Elevator {
	c_responseChan := make(chan T_Elevator)
	ops.C_getElevator <- c_responseChan // Send the response channel to the NodeOperationManager
	elevator := <-c_responseChan        // Receive the connected nodes from the response channel
	return elevator
}
func F_SetElevator(ops T_ElevatorOperations, elevator T_Elevator) {
	ops.C_setElevator <- elevator // Send the connectedNodes directly to be written
}
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
					fmt.Println("Ended GetSet goroutine of NI because of deadlock")
					break WAITFORINTERFACE
				}
			}
		}
	}
}

func F_shouldStop(elevator T_Elevator) bool {
	return (elevator.P_info.Floor == elevator.P_serveRequest.Floor)
}

// her sender jeg ut (fiks deadlock)
// COMMENT: Enig her, funksjonen heter det den skal gjøre
func F_clearRequest(elevator T_Elevator) T_Elevator {
	elevator.P_serveRequest = nil
	elevator.P_info.State = DOOROPEN
	elevator.P_serveRequest = nil
	elevator.P_info.State = DOOROPEN
	return elevator
}

func F_SetElevatorDirection(elevator T_Elevator) T_Elevator { //ta inn requesten og ikke elevator her?
	if elevator.P_serveRequest == nil {
		return elevator
	} else if elevator.StopButton {
		elevator.P_info.Direction = NONE
		F_SetMotorDirection(NONE)

	} else if elevator.P_serveRequest.Floor > elevator.P_info.Floor {
		elevator.P_info.State = MOVING
		elevator.P_info.Direction = UP
		F_SetMotorDirection(UP)

	} else if elevator.P_serveRequest.Floor < elevator.P_info.Floor {
		elevator.P_info.State = MOVING
		elevator.P_info.Direction = DOWN
		F_SetMotorDirection(DOWN)

	} else {
		elevator.P_info.Direction = NONE
		F_SetMotorDirection(NONE)
		elevator = F_clearRequest(elevator)
	}
	return elevator
}
