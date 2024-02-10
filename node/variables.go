package node

var FLOORS int = 4

type T_Node struct {
	Priority       int
	Role           T_NodeRole
	ConnectedNodes []*T_Node
	Elevator       *T_Elevator
}

type T_NodeRole struct {
	Master T_MasterNode
	Slave  T_SlaveNode
}

type T_MasterNode struct {
	GlobalQueue *T_GlobalQueue
}

type T_SlaveNode struct {
	LocalQueue *T_LocalQueue //NB, takes only one order
}

type T_GlobalQueue struct {
	Request T_Request
}

type T_LocalQueue struct {
}

type T_Elevator struct {
	//RequestsToDistribution chan *T_Request
	//RequestsToService      chan *T_Request
	Floor     int
	Direction T_Direction
	Avalibale bool //Thoggled whenever disconnected/unavalebale/door sensor
}

type T_Request struct {
	Calltype   T_Call
	P_Elevator *T_Elevator
	Floor      int
	Direction  T_Direction
}

type T_Call int
type T_Direction int

const (
	Down T_Direction = 0
	Up   T_Direction = 1
)

const (
	Cab  T_Call = 0
	Hall T_Call = 1
)
