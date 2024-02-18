package elevator

type T_Request struct {
	Calltype  T_Call
	Floor     int
	Direction T_Direction //keep for further improvement
}

type T_Call int
type T_Direction int

const (
	CAB  T_Call = 0
	HALL T_Call = 1
)
