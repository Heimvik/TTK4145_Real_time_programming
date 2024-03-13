package node

import (
	"the-elevator/node/elevator"
)

type T_PBNodeRole uint8
type T_MSNodeRole uint8

type T_Node struct {
	NodeInfo       T_NodeInfo
	PBRole         T_PBNodeRole
	GlobalQueue    []T_GlobalQueueEntry
	ConnectedNodes []T_NodeInfo
	Elevator       elevator.T_Elevator
}

type T_NodeInfo struct {
	PRIORITY            uint8
	MSRole              T_MSNodeRole
	TimeUntilDisconnect int
	ElevatorInfo        elevator.T_ElevatorInfo
}

type T_GlobalQueueEntry struct {
	Request           elevator.T_Request
	RequestedNode     uint8
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
	GlobalQueue []T_GlobalQueueEntry
}
type T_SlaveMessage struct {
	Transmitter T_NodeInfo
	Entry       T_GlobalQueueEntry
}

type T_AssignState int

type T_Config struct {
	SlavePort                int     `json:"slaveport"`
	MasterPort               int     `json:"masterport"`
	ElevatorPort             int     `json:"elevatorport"`
	Priority                 uint8   `json:"priority"`
	Nodes                    uint8   `json:"nodes"`
	Floors                   int8    `json:"floors"`
	ReassignPeriod           uint8   `json:"reassignperiod"`
	ConnectionPeriod         int     `json:"connectionperiod"`
	ImmobilePeriod           float64 `json:"immobileperiod"`
	SendPeriod               int     `json:"sendperiod"`
	GetSetPeriod             int     `json:"getsetperiod"`
	AssignBreakoutPeriod     int     `json:"assignbreakoutperiod"`
	MostResponsivePeriod     int     `json:"mostresponsiveperiod"`
	MiddleResponsivePeriod   int     `json:"middleresponsiveperiod"`
	LeastResponsivePeriod    int     `json:"leastresponsiveperiod"`
	TerminationPeriod        int     `json:"terminationperiod"`
	MaxAllowedElevatorErrors int     `json:"maxallowedelevatorerrors"`
	MaxAllowedNodeErrors     int     `json:"maxallowednodeerrors"`
}

const (
	MSROLE_MASTER T_MSNodeRole = 0
	MSROLE_SLAVE  T_MSNodeRole = 1
)
const (
	PBROLE_BACKUP  T_PBNodeRole = 0
	PBROLE_PRIMARY T_PBNodeRole = 1
)
const (
	ASSIGNSTATE_ASSIGN     T_AssignState = 0
	ASSIGNSTATE_WAITFORACK T_AssignState = 1
)

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
}

var ThisNode T_Node

var FLOORS int8
var REASSIGN_PERIOD uint8
var CONNECTION_PERIOD int
var IMMOBILE_PERIOD float64
var SEND_PERIOD int
var GETSET_PERIOD int
var SLAVE_PORT int
var MASTER_PORT int
var ELEVATOR_PORT int
var ASSIGN_BREAKOUT_PERIOD int
var MOST_RESPONSIVE_PERIOD int
var MEDIUM_RESPONSIVE_PERIOD int
var LEAST_RESPONSIVE_PERIOD int
var TERMINATION_PERIOD int
var MAX_ALLOWED_ELEVATOR_ERRORS int
var MAX_ALLOWED_NODE_ERRORS int
