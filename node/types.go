package node

//common include packages:

var FLOORS int
var IP string
var PORT int

type T_Node struct {
	PRIORITY       int //decides role of node based on other alive nodes
	Role           T_NodeRole //role of node
	GlobalQueue    *T_GlobalQueue 
	LocalQueue     *T_LocalQueue
	ConnectedNodes []*T_Node 
	ELEVATOR       *T_Elevator
}

type T_NodeRole int

type T_GlobalQueue struct {
	Request T_Request
}
type T_LocalQueue struct {
}

type T_Elevator struct {
	//Priority int
	//RequestsToDistribution chan *T_Request
	//RequestsToService      chan *T_Request
	Floor     int
	Direction T_Direction
	State     T_ElevatorState
	Avalibale bool //Thoggled whenever disconnected/unavalebale/door sensor
}

type T_ElevatorState int

type T_Request struct {
	Calltype   T_Call
	P_Elevator *T_Elevator
	Floor      int
	Direction  T_Direction
}

type T_Message struct {
	Transmitter T_Node
	TestStr     string
	//Receiver T_Node //In case of FSM
	//MasterMessage T_MasterMessage
	//SlaveMessage  T_SlaveMessage
	//checksum int
}

type T_MasterMessage struct {
	Exist       bool
	GlobalQueue T_GlobalQueue
}
type T_SlaveMessage struct {
	exist bool
}

type T_Config struct {
	Ip       string `json:"host"`
	Port     int    `json:"port"`
	Priority int    `json:"debug"`
	Nodes    int    `json:"nodes"`
	Floors   int    `json:"floors"`
}

type T_Call int
type T_Direction int

const (
	Idle     T_ElevatorState = 0
	Running  T_ElevatorState = 1
	DoorOpen T_ElevatorState = 2
)

const (
	Down T_Direction = 0
	Up   T_Direction = 1
	None T_Direction = 2
)

const (
	Master T_NodeRole = 0
	Slave  T_NodeRole = 1
)

const (
	Cab  T_Call = 0
	Hall T_Call = 1
)
