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
