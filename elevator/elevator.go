package elevator

import (
	"time"

)

type T_ElevatorState int
type T_ElevatorDirection int

const (
	IDLE     T_ElevatorState = iota
	DOOROPEN T_ElevatorState = iota
	MOVING   T_ElevatorState = iota
)

const (
	UP   T_ElevatorDirection = 1
	DOWN T_ElevatorDirection = -1
	NONE T_ElevatorDirection = 0
)

//Keeping this in case of future improvements regarding secondary requirements,
//see single-elevator/elevator.go for inspiration

type T_Elevator struct {
	P_info              *T_ElevatorInfo     //Poitner to the info of this elevator
	P_serveRequest      *T_Request          //Pointer to the current request you are serviceing
	C_receiveRequest    chan T_Request      //Request to put in ServeRequest and do, you will get this from node
	C_distributeRequest chan T_Request      //Requests to distriburte to node, you shoud provide this to node
	C_distributeInfo    chan T_ElevatorInfo //Info on elevator whereabouts, you should provide this to node
	//three last ones has to be channels as elevator and node has seperate goroutines, has to be like this
	//on comilation of one request: redistribute it on D
}
type T_ElevatorInfo struct {
	Direction T_ElevatorDirection
	Floor     int
	State     T_ElevatorState
}

//func for adding request, or create chan which sends to

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
		SetButtonLamp(BT_Cab, elevator.P_serveRequest.Floor, false)
	} else if elevator.P_serveRequest.Calltype == HALL {
		SetButtonLamp(BT_HallDown, elevator.P_serveRequest.Floor, false)
		SetButtonLamp(BT_HallUp, elevator.P_serveRequest.Floor, false)
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
		if elevator.P_serveRequest.Floor > elevator.P_info.Floor {
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