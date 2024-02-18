package node

import (
	"the-elevator/elevator"
	"time"
)

//common include packages:

var FLOORS int
var IP string
var PORT int

type T_Node struct {
	P_info           *T_NodeInfo //role of node
	GlobalQueue    []*T_GlobalQueueEntry
	ConnectedNodes []*T_NodeInfo
	P_ELEVATOR       *elevator.T_Elevator
}
type T_NodeInfo struct {
	PRIORITY int
	Role     T_NodeRole
}

type T_NodeRole int

type T_GlobalQueueEntry struct {
	Request           elevator.T_Request
	RequestedNode     T_NodeInfo //The elevator that got the request
	AssignedNode      T_NodeInfo
	TimeUntilReassign time.Timer
}

type T_MasterMessage struct {
	Transmitter T_NodeInfo
	Receiver    T_NodeInfo //For checking
	GlobalQueue []T_GlobalQueueEntry
	//Checksum int
}
type T_SlaveMessage struct {
	Transmitter  T_NodeInfo
	Receiver     T_NodeInfo //For checking
	RequestInfo  T_GlobalQueueEntry
	ElevatorInfo elevator.T_ElevatorInfo
	//Checksum int
}

type T_Config struct {
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
	Priority int    `json:"priority"`
	Nodes    int    `json:"nodes"`
	Floors   int    `json:"floors"`
}

const (
	Master T_NodeRole = 0
	Slave  T_NodeRole = 1
)
