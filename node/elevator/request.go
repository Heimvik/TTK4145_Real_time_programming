package elevator

var LocalQueue []T_Request

type T_RequestState uint8
type T_Request struct {
	Id        uint16
	State     T_RequestState
	Calltype  T_Call
	Floor     int8
	Direction T_ElevatorDirection //keep for further improvement
}

type T_Call uint8

const (
	NONECALL T_Call = 0
	CAB      T_Call = 1
	HALL     T_Call = 2
)
const (
	UNASSIGNED T_RequestState = 0
	ASSIGNED   T_RequestState = 1
	ACTIVE     T_RequestState = 2
	DONE       T_RequestState = 3
)

func F_ConvertRequestToButtonType(request T_Request) T_ButtonType {
	if request.Calltype == HALL {
		if request.Direction == UP {
			return BT_HallUp
		} else if request.Direction == DOWN {
			return BT_HallDown
		}
	} else if request.Calltype == CAB {
		return BT_Cab
	}
	return 0
}

func F_ConvertButtonTypeToRequest(buttonEvent T_ButtonEvent) T_Request {
	if buttonEvent.Button == BT_HallUp {
		return T_Request{Floor: int8(buttonEvent.Floor), Calltype: HALL, Direction: UP}
	} else if buttonEvent.Button == BT_HallDown {
		return T_Request{Floor: int8(buttonEvent.Floor), Calltype: HALL, Direction: DOWN}
	} else if buttonEvent.Button == BT_Cab {
		return T_Request{Floor: int8(buttonEvent.Floor), Calltype: CAB, Direction: NONE}
	}
	return T_Request{}
}

func F_ReceiveRequest(req T_Request, elevator T_Elevator) T_Elevator {
	elevator.P_serveRequest = &req
	elevator.P_serveRequest.State = ACTIVE
	return elevator
}

func F_ClearRequest(elevator T_Elevator) T_Elevator {
	elevator.P_serveRequest = nil
	elevator.P_info.State = DOOROPEN
	return elevator
}

// COMMENT: Legg ut i Run_elevator
func F_SendRequest(button T_ButtonEvent, requestOut chan T_Request, elevator T_Elevator) {
	request := F_ConvertButtonTypeToRequest(button)
	if button.Button == BT_Cab {
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: request.Calltype, Floor: request.Floor, Direction: request.Direction}
	} else {
		requestOut <- T_Request{Id: uint16(elevator.CurrentID), State: 0, Calltype: request.Calltype, Floor: request.Floor, Direction: request.Direction}
	}
}
