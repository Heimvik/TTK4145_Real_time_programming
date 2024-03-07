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

type T_ElevatorOperations struct {
	C_readElevator         chan chan T_Elevator
	C_writeElevator        chan T_Elevator
	C_readAndWriteElevator chan chan T_Elevator
}

type T_ElevatorChannels struct {
	C_readElevator  chan T_Elevator
	C_writeElevator chan T_Elevator
	C_quitGetSet          chan bool
	// C_timerStart  chan bool
	C_timerStop chan bool
	C_timerTimeout chan bool
	C_buttonPress chan T_ButtonEvent
	C_floorSensor chan int
	C_obstr       chan bool
	C_stop        chan bool
}

func F_InitElevatorChannels() T_ElevatorChannels {
	return T_ElevatorChannels{
		C_readElevator:  make(chan T_Elevator),
		C_writeElevator: make(chan T_Elevator),
		C_quit:          make(chan bool),
		// C_timerStart:  make(chan bool),
		C_timerStop: make(chan bool),
		C_timerTimeout: make(chan bool),
		C_buttonPress: make(chan T_ButtonEvent),
		C_floorSensor: make(chan int),
		C_obstr:       make(chan bool),
		C_stop:        make(chan bool),
	}
}

func F_GetElevator(ops T_ElevatorOperations) T_Elevator {
	c_responseChan := make(chan T_Elevator)
	ops.C_readElevator <- c_responseChan // Send the response channel to the NodeOperationManager
	elevator := <-c_responseChan         // Receive the connected nodes from the response channel
	return elevator
}
func F_SetElevator(ops T_ElevatorOperations, elevator T_Elevator) {
	ops.C_writeElevator <- elevator // Send the connectedNodes directly to be written
}
func F_GetAndSetElevator(ops T_ElevatorOperations, c_readElevator chan T_Elevator, c_writeElevator chan T_Elevator, c_quit chan bool) { //let run in a sepreate goroutine
	getSetTimer := time.NewTicker(time.Duration(2) * time.Second) //Hardkode 2 inntil videre, sync med initfil
	c_responsChan := make(chan T_Elevator)

	ops.C_readAndWriteElevator <- c_responsChan
	for {
		select {
		case oldElevator := <-c_responsChan:
			c_readElevator <- oldElevator
		case newElevator := <-c_writeElevator:
			c_responsChan <- newElevator
		case <-c_quit:
			return
		case <-getSetTimer.C:
			fmt.Println("Elevator deadlock upon getset")
			//F_WriteLog("Ended GetSet goroutine of CN because of deadlock")
		}
	}
}

func F_ShouldElevatorStop(elevator T_Elevator) bool {
	return ((elevator.P_info.State == MOVING) && (elevator.P_info.Floor == elevator.P_serveRequest.Floor)) || elevator.StopButton
}

// her sender jeg ut (fiks deadlock)
// COMMENT: Enig her, funksjonen heter det den skal gjøre
func F_clearRequest(elevator T_Elevator) T_Elevator {
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
