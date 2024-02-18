package node

import (
	"the-elevator/elevator"
	"time"
)

//common include packages:

type T_Node struct {
	P_info         *T_NodeInfo //role of node
	GlobalQueue    []T_GlobalQueueEntry
	ConnectedNodes []T_NodeInfo
	P_ELEVATOR     *elevator.T_Elevator
}
type T_NodeRole int
type T_NodeInfo struct {
	PRIORITY     int
	Role         T_NodeRole
	ElevatorInfo elevator.T_ElevatorInfo
}

type T_GlobalQueueEntry struct {
	Id                int
	Request           elevator.T_Request
	RequestedNode     T_NodeInfo //The elevator that got the request
	AssignedNode      T_NodeInfo
	TimeUntilReassign time.Timer
}

type Communication interface {
	I_Transmit()
	I_Receive()
}

type T_MasterMessage struct {
	Transmitter T_NodeInfo
	Receiver    T_NodeInfo //For checking
	GlobalQueue []T_GlobalQueueEntry
	//Checksum int
}
type T_SlaveMessage struct {
	Transmitter T_NodeInfo
	Receiver    T_NodeInfo         //For checking
	RequestInfo T_GlobalQueueEntry //find a better name?
	//Checksum int
}

type T_Config struct {
	Ip           string `json:"ip"`
	Port         int    `json:"port"`
	Priority     int    `json:"priority"`
	Nodes        int    `json:"nodes"`
	Floors       int    `json:"floors"`
	ReassignTime int    `json:"reassigntime"`
}

const (
	MASTER T_NodeRole = 0
	SLAVE  T_NodeRole = 1
)

// Global Variables
var ThisNode T_Node

var FLOORS int
var IP string
var PORT int
var REASSIGNTIME int
