package elevator

var LocalQueue []T_Request

type T_RequestState int
type T_Request struct {
	Id        int
	State     T_RequestState
	Calltype  T_Call
	Floor     int
	Direction T_ElevatorDirection //keep for further improvement
}

type T_Call int

const (
	NONECALL T_Call = 0
	CAB      T_Call = 1
	HALL     T_Call = 2
)
const (
	UNASSIGNED T_RequestState = 0
	ASSIGNED   T_RequestState = 1
	DONE       T_RequestState = 2
)
