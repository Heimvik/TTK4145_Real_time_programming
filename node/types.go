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
	ConnectedNodes []int
	ELEVATOR       *T_Elevator
}

type T_NodeRole int

type T_GlobalQueue struct {
	Request T_Request
}
type T_LocalQueue struct {
}


type T_Message struct {
	Transmitter int
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

const (
	Master T_NodeRole = 0
	Slave  T_NodeRole = 1
)
