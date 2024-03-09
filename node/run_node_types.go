package node

import (
	"the-elevator/node/elevator"
)

//common include packages:

type T_Node struct {
	NodeInfo       T_NodeInfo //role of node
	GlobalQueue    []T_GlobalQueueEntry
	ConnectedNodes []T_NodeInfo
	Elevator       elevator.T_Elevator //Its info needs to point at NodeInfo.ElevatorInfo
}
type T_MasterSlaveRole uint8
type T_PrimaryBackupRole uint8

type T_NodeInfo struct {
	PRIORITY            uint8
	MSRole              T_MasterSlaveRole
	PBRole				T_PrimaryBackupRole
	TimeUntilDisconnect int
	ElevatorInfo        elevator.T_ElevatorInfo
}

type T_GlobalQueueEntry struct {
	Request           elevator.T_Request
	RequestedNode     uint8 //PRIORITY of the one that got request
	AssignedNode      uint8
	TimeUntilReassign uint8
}

type T_AckObject struct {
	ObjectToAcknowledge        interface{}
	ObjectToSupportAcknowledge interface{}
	C_Acknowledgement          chan bool
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

type T_AssignState int

type T_Config struct {
	Ip                     string `json:"ip"`
	SlavePort              int    `json:"slaveport"`
	MasterPort             int    `json:"masterport"`
	ElevatorPort           int    `json:"elevatorport"`
	Priority               uint8  `json:"priority"`
	Nodes                  uint8  `json:"nodes"`
	Floors                 int8   `json:"floors"`
	ReassignTime           uint8  `json:"reassigntime"`
	ConnectionTime         int    `json:"connectiontime"`
	SendPeriod             int    `json:"sendperiod"`
	GetSetPeriod           int    `json:"getsetperiod"`
	AssignBreakoutPeriod   int    `json:"assignbreakoutperiod"`
	MostResponsivePeriod   int    `json:"mostresponsiveperiod"`
	MiddleResponsivePeriod int    `json:"middleresponsiveperiod"`
	LeastResponsivePeriod  int    `json:"leastresponsiveperiod"`
}

const (
	BACKUP T_PrimaryBackupRole = 0
	PRIMARY T_PrimaryBackupRole = 1
)

const (
	MASTER T_MasterSlaveRole = 0
	SLAVE  T_MasterSlaveRole = 1
)
const (
	ASSIGN     T_AssignState = 0
	WAITFORACK T_AssignState = 1
)

// NodeOperation represents an operation to be performed on T_Node
type T_GetSetNodeInfoInterface struct {
	c_get chan T_NodeInfo
	c_set chan T_NodeInfo
}
type T_GetSetGlobalQueueInterface struct {
	c_get chan []T_GlobalQueueEntry
	c_set chan []T_GlobalQueueEntry
}
type T_GetSetConnectedNodesInterface struct {
	c_get chan []T_NodeInfo
	c_set chan []T_NodeInfo
}

type T_NodeOperations struct {
	c_getNodeInfo    chan chan T_NodeInfo
	c_setNodeInfo    chan T_NodeInfo
	c_getSetNodeInfo chan chan T_NodeInfo

	c_getGlobalQueue    chan chan []T_GlobalQueueEntry
	c_setGlobalQueue    chan []T_GlobalQueueEntry
	c_getSetGlobalQueue chan chan []T_GlobalQueueEntry

	c_getConnectedNodes    chan chan []T_NodeInfo
	c_setConnectedNodes    chan []T_NodeInfo
	c_getSetConnectedNodes chan chan []T_NodeInfo
	// Add more channels for other operations as needed
}

// Global Variables
var ThisNode T_Node

var FLOORS int8
var IP string
var REASSIGNTIME uint8
var CONNECTIONTIME int
var SENDPERIOD int
var GETSETPERIOD int
var SLAVEPORT int
var MASTERPORT int
var ELEVATORPORT int
var ASSIGNBREAKOUTPERIOD int
var MOSTRESPONSIVEPERIOD int
var MEDIUMRESPONSIVEPERIOD int
var LEASTRESPONSIVEPERIOD int
