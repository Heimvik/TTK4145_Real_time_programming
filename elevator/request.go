package elevator

var LocalQueue []T_Request

type T_Request struct {
	Calltype   T_Call
	Floor      int
	Direction  T_Direction //keep for further improvement
}


type T_Call int
type T_Direction int

const (
	Cab  T_Call = 0
	Hall T_Call = 1
)