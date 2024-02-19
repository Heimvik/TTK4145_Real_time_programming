package node

import (
	"the-elevator/elevator"
)

//common include packages:

type T_Node struct {
	Info           T_NodeInfo //role of node
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
type T_EntryState int
type T_GlobalQueueEntry struct {
	Id                int
	State             T_EntryState
	Request           elevator.T_Request
	RequestedNode     T_NodeInfo //The elevator that got the request
	AssignedNode      T_NodeInfo
	TimeUntilReassign int
}

type T_MasterMessage struct {
	Transmitter T_NodeInfo
	//Receiver    T_NodeInfo //For checking
	GlobalQueue []T_GlobalQueueEntry
	//Checksum int
}
type T_SlaveMessage struct {
	Transmitter T_NodeInfo
	//Receiver    T_NodeInfo         //For checking
	Entry T_GlobalQueueEntry //find a better name?
	//Checksum int
}

type T_Config struct {
	Ip           string `json:"ip"`
	SlavePort    int    `json:"slaveport"`
	MasterPort   int    `json:"masterport"`
	Priority     int    `json:"priority"`
	Nodes        int    `json:"nodes"`
	Floors       int    `json:"floors"`
	ReassignTime int    `json:"reassigntime"`
	MMMills      int    `json:"mmmills"`
}

const (
	MASTER T_NodeRole = 0
	SLAVE  T_NodeRole = 1
)
const (
	UNASSIGNED T_EntryState = 0
	ASSIGNED   T_EntryState = 1
	DONE       T_EntryState = 2
)

// NodeOperation represents an operation to be performed on T_Node
type T_NodeOperations struct {
	c_readNodeInfo        chan chan T_NodeInfo
	c_writeNodeInfo       chan T_NodeInfo
	c_readGlobalQueue     chan chan []T_GlobalQueueEntry
	c_writeGlobalQueue    chan []T_GlobalQueueEntry
	c_readConnectedNodes  chan chan []T_NodeInfo
	c_writeConnectedNodes chan []T_NodeInfo
	c_readElevator        chan chan *elevator.T_Elevator
	// Add more channels for other operations as needed
}

// Global Variables
var ThisNode T_Node

var FLOORS int
var IP string
var REASSIGNTIME int
var MMMILLS int
var SLAVEPORT int
var MASTERPORT int