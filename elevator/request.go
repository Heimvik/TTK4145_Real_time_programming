package elevator

type T_Request struct {
	Calltype   T_Call
	P_Elevator *T_Elevator
	Floor      int
	// Direction  T_Direction //keep for further improvement
}


type T_Call int

const (
	Cab  T_Call = 0
	Hall T_Call = 1
)