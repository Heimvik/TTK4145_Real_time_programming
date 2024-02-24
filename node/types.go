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
	PRIORITY            int
	Role                T_NodeRole
	ElevatorInfo        elevator.T_ElevatorInfo
	TimeUntilDisconnect float32
}

type T_GlobalQueueEntry struct {
	Request           elevator.T_Request
	RequestedNode     T_NodeInfo //The elevator that got the request
	AssignedNode      T_NodeInfo
	TimeUntilReassign float32
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
	Ip             string  `json:"ip"`
	SlavePort      int     `json:"slaveport"`
	MasterPort     int     `json:"masterport"`
	ElevatorPort   int     `json:"elevatorport"`
	Priority       int     `json:"priority"`
	Nodes          int     `json:"nodes"`
	Floors         int     `json:"floors"`
	ReassignTime   float32 `json:"reassigntime"`
	ConnectionTime float32 `json:"connectiontime"`
	SendPeriod     int     `json:"sendperiod"`
	GetSetPeriod   int     `json:"getsetperiod"`
}

const (
	MASTER T_NodeRole = 0
	SLAVE  T_NodeRole = 1
)

// NodeOperation represents an operation to be performed on T_Node
type T_NodeOperations struct {
	c_readNodeInfo         chan chan T_NodeInfo
	c_writeNodeInfo        chan T_NodeInfo
	c_readAndWriteNodeInfo chan chan T_NodeInfo

	c_readGlobalQueue         chan chan []T_GlobalQueueEntry
	c_writeGlobalQueue        chan []T_GlobalQueueEntry
	c_readAndWriteGlobalQueue chan chan []T_GlobalQueueEntry

	c_readConnectedNodes         chan chan []T_NodeInfo
	c_writeConnectedNodes        chan []T_NodeInfo
	c_readAndWriteConnectedNodes chan chan []T_NodeInfo

	c_readElevator         chan chan elevator.T_Elevator
	c_writeElevator        chan elevator.T_Elevator
	c_readAndWriteElevator chan chan elevator.T_Elevator
	// Add more channels for other operations as needed
}

// Global Variables
var ThisNode T_Node

var FLOORS int
var IP string
var REASSIGNTIME float32
var CONNECTIONTIME float32
var SENDPERIOD int
var GETSETPERIOD int
var SLAVEPORT int
var MASTERPORT int
var ELEVATORPORT int
