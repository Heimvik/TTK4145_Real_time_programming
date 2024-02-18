package elevator

var LocalQueue []T_Request

type T_Request struct {
	Calltype   T_Call
	Floor      int
	Direction  T_ElevatorDirection //keep for further improvement
}

type T_Call int


const (
	CAB  T_Call = 0
	HALL T_Call = 1
)

