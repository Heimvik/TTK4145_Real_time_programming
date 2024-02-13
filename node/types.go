package node

import (
	"the-elevator/elevator"
)

//common include packages:

var FLOORS int
var IP string
var PORT int

type T_Node struct {
	Info *T_NodeInfo //role of node
	//GlobalQueue    *T_GlobalQueue
	ConnectedNodes []*T_NodeInfo
	//ELEVATOR *elevator.T_Elevator
}
type T_NodeInfo struct {
	PRIORITY int
	Role     T_NodeRole
}

type T_NodeRole int

type T_GlobalQueue struct {
	Request elevator.T_Request
}
type T_LocalQueue struct {
}

type T_Message struct {
	Transmitter T_NodeInfo
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
