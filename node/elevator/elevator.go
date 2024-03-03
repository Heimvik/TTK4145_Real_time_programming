package elevator

import (
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
	CurrentID           int
	P_info              *T_ElevatorInfo     //MUST be pointer to info (points to info stored in ThisNode.NodeInfo.ElevatorInfo)
	P_serveRequest      *T_Request          //Pointer to the current request you are serviceing
	C_receiveRequest    chan T_Request      //Request to put in ServeRequest and do, you will get this from node
	C_distributeRequest chan T_Request      //Requests to distriburte to node, you shoud provide this to node
	C_distributeInfo    chan T_ElevatorInfo //Info on elevator whereabouts, you should provide this to node
	//three last ones has to be channels as elevator and node has seperate goroutines, has to be like this
	//on comilation of one request: redistribute it on D
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
			//F_WriteLog("Ended GetSet goroutine of CN because of deadlock")
		}
	}
}

func Init_Elevator(requestIn chan T_Request, requestOut chan T_Request) T_Elevator {
	return T_Elevator{
		P_info:              &T_ElevatorInfo{Direction: NONE, Floor: 0, State: IDLE},
		P_serveRequest:      nil,
		C_receiveRequest:    requestIn,
		C_distributeRequest: requestOut,
		C_distributeInfo:    make(chan T_ElevatorInfo),
	}
}

func F_shouldStop(elevator T_Elevator) bool {
	if elevator.P_info.State == MOVING {
		if elevator.P_info.Floor == elevator.P_serveRequest.Floor {
			return true
		}
	}
	return false
}

func F_clearRequest(elevator T_Elevator) {
	//sjekk at det er en request å cleare
	if elevator.P_serveRequest == nil {
		return
	} else if elevator.P_serveRequest.Calltype == CAB { //skru av lys
		SetButtonLamp(BT_Cab, int(elevator.P_serveRequest.Floor), false)
	} else if elevator.P_serveRequest.Calltype == HALL {
		SetButtonLamp(BT_HallDown, int(elevator.P_serveRequest.Floor), false)
		SetButtonLamp(BT_HallUp, int(elevator.P_serveRequest.Floor), false)
	}
	//set request til done
	elevator.P_serveRequest.State = DONE
	// elevator.C_distributeRequest <- *elevator.P_serveRequest
	SetMotorDirection(MD_Stop)
	Elevator.P_info.State = DOOROPEN
	SetDoorOpenLamp(true)
	time.Sleep(3 * time.Second) //placeholder
	SetDoorOpenLamp(false)
	Elevator.P_info.State = IDLE
	elevator.P_serveRequest = nil
}

func F_chooseDirection(elevator T_Elevator) {
	if elevator.P_serveRequest == nil {
		return
	} else if C_stop {

		SetMotorDirection(MD_Stop)
		elevator.P_info.State = IDLE
		elevator.P_info.Direction = NONE

	} else if elevator.P_serveRequest.Floor > elevator.P_info.Floor {
		elevator.P_info.Direction = UP
		elevator.P_info.State = MOVING
		SetMotorDirection(MD_Up)

	} else if elevator.P_serveRequest.Floor < elevator.P_info.Floor {
		elevator.P_info.Direction = DOWN
		elevator.P_info.State = MOVING
		SetMotorDirection(MD_Down)
	} else {
		elevator.P_info.Direction = NONE
		F_clearRequest(elevator)
	}
}
